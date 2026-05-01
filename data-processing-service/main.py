import io
import re
import json
import difflib
import cv2
import numpy as np
import pandas as pd
import pytesseract
import networkx as nx
import geopandas as gpd
from pdf2image import convert_from_bytes
from shapely.geometry import Polygon
from shapely.ops import polygonize, unary_union
from fastapi import FastAPI, UploadFile, File, HTTPException
from pydantic import BaseModel
from typing import List
import city2graph as c2g

# Import your city2graph module here
# import city2graph 

app = FastAPI(title="Field Scheduler Parser Service")

@app.get("/health")
async def health_check():
    return {"status": "ok"}

# ==========================================
# 1. PDF SCHEDULE PARSING & OCR
# ==========================================

def enhance_cell_for_ocr(cell_crop):
    cell_crop = cv2.resize(cell_crop, None, fx=2.5, fy=2.5, interpolation=cv2.INTER_CUBIC)
    gray = cv2.cvtColor(cell_crop, cv2.COLOR_BGR2GRAY)
    _, thresh = cv2.threshold(gray, 0, 255, cv2.THRESH_BINARY | cv2.THRESH_OTSU)
    padded_cell = cv2.copyMakeBorder(thresh, 20, 20, 20, 20, cv2.BORDER_CONSTANT, value=[255, 255, 255])
    return padded_cell

def reconstruct_grid_and_parse(image):
    img = cv2.cvtColor(np.array(image), cv2.COLOR_RGB2BGR)
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    
    _, thresh = cv2.threshold(gray, 128, 255, cv2.THRESH_BINARY_INV | cv2.THRESH_OTSU)

    kernel_length = np.array(gray).shape[1] // 40
    vert_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (1, kernel_length))
    hori_kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (kernel_length, 1))

    img_temp1 = cv2.erode(thresh, vert_kernel, iterations=3)
    vert_lines = cv2.dilate(img_temp1, vert_kernel, iterations=3)

    img_temp2 = cv2.erode(thresh, hori_kernel, iterations=3)
    hori_lines = cv2.dilate(img_temp2, hori_kernel, iterations=3)

    table_mask = cv2.addWeighted(vert_lines, 0.5, hori_lines, 0.5, 0.0)
    _, table_mask = cv2.threshold(table_mask, 50, 255, cv2.THRESH_BINARY)

    contours, _ = cv2.findContours(table_mask, cv2.RETR_TREE, cv2.CHAIN_APPROX_SIMPLE)
    bounding_boxes =[cv2.boundingRect(c) for c in contours]

    raw_x, raw_y = [],[]
    for x, y, w, h in bounding_boxes:
        if w > 40 and h > 15 and w < img.shape[1] * 0.9:
            raw_x.extend([x, x + w])
            raw_y.extend([y, y + h])

    def cluster_coords(coords, threshold=10):
        if not coords: return[]
        coords = sorted(list(set(coords)))
        clustered = [coords[0]]
        for c in coords[1:]:
            if c - clustered[-1] > threshold:
                clustered.append(c)
        return clustered

    cols = cluster_coords(raw_x)
    rows = cluster_coords(raw_y)

    extracted_data =[]
    skip_cells = set()
    margin = 5

    for i in range(len(rows) - 1):
        row_data =[]
        for j in range(len(cols) - 1):
            if (i, j) in skip_cells:
                row_data.append("") 
                continue
                
            x1, x2 = cols[j] + margin, cols[j+1] - margin
            y1 = rows[i] + margin
            
            current_i = i
            while current_i < len(rows) - 2:
                y_bottom = rows[current_i + 1]
                roi_width_padding = int((x2 - x1) * 0.1)
                border_check_area = hori_lines[y_bottom-3 : y_bottom+3, x1+roi_width_padding : x2-roi_width_padding]
                
                if np.sum(border_check_area) > 0:
                    break
                current_i += 1
                skip_cells.add((current_i, j))
                
            y2 = rows[current_i + 1] - margin
            cell_crop = img[y1:y2, x1:x2]
            enhanced_cell = enhance_cell_for_ocr(cell_crop)
            
            text = pytesseract.image_to_string(enhanced_cell, config='--psm 6').strip()
            row_data.append(text)
            
        extracted_data.append(row_data)

    return extracted_data

def detect_table_structure_from_pdf_bytes(pdf_bytes: bytes):
    images = convert_from_bytes(pdf_bytes)
    entries =[]
    for i, image in enumerate(images):
        parsed_table = reconstruct_grid_and_parse(image)
        if len(parsed_table) == 0:
            continue
            
        if i == 0:
            entries.append(parsed_table[0])  # Add header row
        for row in parsed_table[1:]:
            entries.append(row)
    return entries

def flatten_schedule_data(data):
    if len(data) <= 1:
        return []
        
    df = pd.DataFrame(data[1:])
    parsed_blocks = []
    current_block =[]

    for index, row in df.iterrows():
        day_val = row[0]
        if pd.notna(day_val) and str(day_val).strip() != "":
            if current_block:
                parsed_blocks.append(current_block)
                current_block =[]
        current_block.append(row.values.tolist())

    if current_block:
        parsed_blocks.append(current_block)

    flat_schedule =[]

    for block in parsed_blocks:
        columns_lists = [[], [], [], [], [],[]]
        
        for row in block:
            for col_idx in range(6):
                val = row[col_idx] if col_idx < len(row) else ""
                if pd.notna(val) and str(val).strip() != "":
                    parts = [p.strip() for p in str(val).split('\n') if p.strip()]
                    columns_lists[col_idx].extend(parts)
                    
        N = max((len(col) for col in columns_lists[0:5]), default=1)
        N = max(1, N)
        
        for i in range(N):
            event = {}
            for col_idx, col_name in enumerate(["Day", "Date", "Time", "Venue", "Conductor", "Remark"]):
                col_list = columns_lists[col_idx]
                
                if len(col_list) == 0:
                    val = ""
                elif len(col_list) == 1:
                    val = col_list[0]
                elif len(col_list) == N:
                    val = col_list[i]
                else:
                    if col_idx == 5 and N == 1:
                        val = " ".join(col_list)
                    else:
                        val = col_list[i] if i < len(col_list) else col_list[-1]
                
                event[col_name] = val
            flat_schedule.append(event)

    return flat_schedule

@app.post("/parse/pdf")
async def parse_pdf(file: UploadFile = File(...)):
    if not file.filename.endswith(".pdf"):
        raise HTTPException(status_code=400, detail="Only PDF files are supported")
    try:
        pdf_bytes = await file.read()
        raw_data = detect_table_structure_from_pdf_bytes(pdf_bytes)
        flat_schedule = flatten_schedule_data(raw_data)
        return flat_schedule
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# ==========================================
# 2. GIS MAP PARSING & TOPOLOGY
# ==========================================

LEKKI_POLYGON = Polygon([
    (3.482269, 6.448432), (3.482871, 6.432581), (3.461505, 6.429083),
    (3.453218, 6.437141), (3.453329, 6.449450), (3.456248, 6.450003),
    (3.460391, 6.447214), (3.469272, 6.448598)
])

MAP_CACHE = {}

def normalize_street_names(name: str) -> list[str]:
    """Extracts core names, handles parentheses, and strips suffixes."""
    if not isinstance(name, str):
        return[]
    
    name = name.lower()
    alt_names = re.findall(r'\((.*?)\)', name)
    base_name = re.sub(r'\(.*?\)', '', name)
    raw_variations = [base_name] + alt_names
    
    cleaned_variations =[]
    suffixes = r'\b(street|st|crescent|cres|drive|dr|road|rd|avenue|ave|close|cl|way|lane|ln|highway|hwy|boulevard|blvd)\b'
    
    for var in raw_variations:
        var = re.sub(r'[^\w\s]', '', var)
        var = re.sub(suffixes, '', var)
        var = re.sub(r'\bthe\b', '', var)
        var = re.sub(r'\s+', ' ', var).strip()
        if var:
            cleaned_variations.append(var)
            
    return cleaned_variations

def my_city2graph_fetch_function(polygon):
    return c2g.load_overture_data(area=polygon, types=["segment"])['segment']

def get_lekki_city2graph():
    """Fetches city2graph data and builds a NetworkX graph."""
    if "graph" in MAP_CACHE:
        return MAP_CACHE["graph"], MAP_CACHE["gdf"]

    try:
        df = my_city2graph_fetch_function(LEKKI_POLYGON)
    except Exception as e:
        raise Exception(f"Failed to fetch city2graph data: {e}")

    G = nx.MultiGraph()
    G.graph['crs'] = "EPSG:4326"
    
    for idx, row in df.iterrows():
        connectors = row.get('connectors')
        geom = row.get('geometry')
        
        names_dict = row.get('names', {})
        primary = row.get('primary_name', '')
        
        all_names = [primary] if primary else[]
        if isinstance(names_dict, dict):
            all_names.extend([v for k, v in names_dict.items() if v])
            
        if isinstance(connectors, list) and len(connectors) >= 2:
            u = connectors[0]['connector_id']
            v = connectors[-1]['connector_id']
            G.add_edge(u, v, key=row['id'], geometry=geom, names=all_names)

    gdf = gpd.GeoDataFrame(df, geometry='geometry', crs="EPSG:4326")

    MAP_CACHE["graph"] = G
    MAP_CACHE["gdf"] = gdf
    
    return G, gdf

class StreetQuery(BaseModel):
    streets: List[str]

@app.post("/fetch/block-by-topology")
async def fetch_block_by_topology(query: StreetQuery):
    try:
        G, _ = get_lekki_city2graph()
        
        query_variants = {}
        for req_street in query.streets:
            query_variants[req_street] = normalize_street_names(req_street)
            
        subgraph_edges =[]
        
        for u, v, key, data in G.edges(keys=True, data=True):
            edge_names_raw = data.get('names', [])
            edge_norms =[]
            for en in edge_names_raw:
                edge_norms.extend(normalize_street_names(en))
            
            if not edge_norms:
                continue
                
            matched = False
            for req_street, req_norms in query_variants.items():
                for rn in req_norms:
                    for en in edge_norms:
                        if rn == en: matched = True
                        elif (len(rn) >= 5 and rn in en) or (len(en) >= 5 and en in rn): matched = True
                        elif len(rn) >= 4 and difflib.SequenceMatcher(None, rn, en).ratio() > 0.85: matched = True
                        
                        if matched:
                            subgraph_edges.append((u, v, key))
                            break
                    if matched: break
                if matched: break

        if not subgraph_edges:
             raise HTTPException(status_code=404, detail="Requested streets not found in graph.")

        H = G.edge_subgraph(subgraph_edges)
        H_undirected = nx.Graph(H)
        
        try:
            cycles = nx.cycle_basis(H_undirected)
        except Exception:
            cycles =[]
            
        if not cycles:
            raise HTTPException(status_code=404, detail="Streets found, but they do not form a closed block.")
            
        best_cycle = None
        best_score = -1
        
        for cycle in cycles:
            cycle_matched_requests = set()
            for i in range(len(cycle)):
                u = cycle[i]
                v = cycle[(i + 1) % len(cycle)]
                
                edge_data = H.get_edge_data(u, v)
                if not edge_data: continue
                
                for key, data in edge_data.items():
                    edge_names_raw = data.get('names',[])
                    edge_norms =[]
                    for en in edge_names_raw:
                        edge_norms.extend(normalize_street_names(en))
                        
                    for req_street, req_norms in query_variants.items():
                        if req_street in cycle_matched_requests: continue
                        matched = False
                        for rn in req_norms:
                            for en in edge_norms:
                                if rn == en: matched = True
                                elif (len(rn) >= 5 and rn in en) or (len(en) >= 5 and en in rn): matched = True
                                elif len(rn) >= 4 and difflib.SequenceMatcher(None, rn, en).ratio() > 0.85: matched = True
                                if matched: break
                            if matched: break
                        if matched:
                            cycle_matched_requests.add(req_street)
            
            score = len(cycle_matched_requests)
            adjusted_score = score - (len(cycle) * 0.01)
            
            if adjusted_score > best_score:
                best_score = adjusted_score
                best_cycle = cycle

        if not best_cycle:
            raise HTTPException(status_code=404, detail="Could not resolve a valid block from the provided streets.")

        cycle_lines =[]
        for i in range(len(best_cycle)):
            u = best_cycle[i]
            v = best_cycle[(i + 1) % len(best_cycle)]
            edge_data = G.get_edge_data(u, v)
            if edge_data:
                first_key = list(edge_data.keys())[0]
                geom = edge_data[first_key].get('geometry')
                if geom: cycle_lines.append(geom)

        polys = list(polygonize(unary_union(cycle_lines)))
        if not polys:
             raise HTTPException(status_code=500, detail="Failed to assemble polygon.")
             
        gs = gpd.GeoSeries([polys[0]], crs="EPSG:4326")
        coords = list(gs.iloc[0].exterior.coords)
        formatted_coords = [[lat, lon] for lon, lat in coords]
        
        return {
            "requested_streets": query.streets,
            "match_score": best_score,
            "coordinates": formatted_coords
        }

    except HTTPException as he:
        raise he
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/fetch/map-blocks")
async def fetch_map_blocks():
    """Generates all valid polygons in the bounding box to render the background map."""
    try:
        _, gdf = get_lekki_city2graph()
        
        merged_lines = unary_union(gdf['geometry'].tolist())
        lines_list = list(merged_lines.geoms) if hasattr(merged_lines, 'geoms') else [merged_lines]
        
        polygons = list(polygonize(lines_list))
        blocks = gpd.GeoDataFrame(geometry=polygons, crs="EPSG:4326")
        
        # Temporary projection to UTM zone 31N (Lagos) for accurate area filtering in sq meters
        blocks_metric = blocks.to_crs("EPSG:32631")
        valid_indices = (blocks_metric.area > 500) & (blocks_metric.area < 500000)
        blocks_filtered = blocks[valid_indices]
        
        output_polygons =[]
        for poly in blocks_filtered.geometry:
            coords = list(poly.exterior.coords)
            formatted_coords = [[lat, lon] for lon, lat in coords]
            output_polygons.append(formatted_coords)
            
        return {
            "blocks_count": len(output_polygons),
            "polygons": output_polygons
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
