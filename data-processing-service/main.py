from fastapi import FastAPI, UploadFile, File, HTTPException
import cv2
import numpy as np
import pytesseract
from pdf2image import convert_from_bytes
import pandas as pd
import osmnx as ox
import networkx as nx

app = FastAPI(title="Field Scheduler Parser Service")

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
        return[]
        
    # data[0] is header, data[1:] are the rows. 
    # We load it directly into pandas instead of going through a CSV file
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

    flat_schedule = []

    for block in parsed_blocks:
        columns_lists = [[],[], [], [], [],[]]
        
        for row in block:
            for col_idx in range(6):
                val = row[col_idx] if col_idx < len(row) else ""
                if pd.notna(val) and str(val).strip() != "":
                    parts =[p.strip() for p in str(val).split('\n') if p.strip()]
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
        # Read the PDF into memory
        pdf_bytes = await file.read()
        
        # Extract raw grid
        raw_data = detect_table_structure_from_pdf_bytes(pdf_bytes)
        
        # Flatten structure into JSON dicts
        flat_schedule = flatten_schedule_data(raw_data)
        
        return flat_schedule
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

LEKKI_ADDRESS = "24 Abike Sulaiman St, Lekki Phase I, Lekki 106104, Lagos"

@app.get("/fetch/map-blocks")
async def fetch_map_blocks(radius: int = 1000):
    """
    Fetches street blocks from OSMnx around the hardcoded Lekki address.
    Returns a list of polygons for the frontend to render.
    """
    try:
        # 1. Download graph
        G = ox.graph_from_address(LEKKI_ADDRESS, dist=radius, network_type='all')
        
        # 2. Convert to an undirected graph to easily find loops
        Gu = ox.utils_graph.get_undirected(G)

        G_simple = nx.Graph(Gu)
        
        # 3. Find all closed loops (cycles) in the network
        # Each cycle is a list of Node IDs that form a block
        cycles = nx.minimum_cycle_basis(G_simple)
        
        output_polygons = []
        
        # 4. Extract coordinates for each block
        for cycle in cycles:
            # A valid block needs at least 3 intersections
            if len(cycle) > 2:
                block_coords = []
                
                for node_id in cycle:
                    # Extract the GPS coordinates directly from the node data
                    lat = Gu.nodes[node_id]['y']
                    lon = Gu.nodes[node_id]['x']
                    
                    # Format as [lat, lon] for Leaflet/Mapbox
                    block_coords.append([lat, lon])
                
                output_polygons.append(block_coords)
                
        return {
            "address": LEKKI_ADDRESS,
            "blocks_count": len(output_polygons),
            "polygons": output_polygons
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "ok"}
