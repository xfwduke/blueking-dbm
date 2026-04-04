#!/usr/bin/env python3
"""
Increase answer space for Q22's three sub-questions in the exam PDF.
Strategy: render page to high-res image, slice at sub-question boundaries,
reassemble with extra whitespace inserted between sub-questions.
"""
import fitz  # PyMuPDF
from PIL import Image
import io

INPUT_PDF = "/home/ubuntu/.cursor/projects/workspace/uploads/page1_left_1_.pdf"
OUTPUT_PDF = "/workspace/page1_left_1_spaced.pdf"

DPI = 300
SCALE = DPI / 72.0

EXTRA_SPACE_Q1 = 80  # extra pts after Q22(1) hint
EXTRA_SPACE_Q2 = 80  # extra pts after Q22(2) text

extra_px_q1 = int(EXTRA_SPACE_Q1 * SCALE)
extra_px_q2 = int(EXTRA_SPACE_Q2 * SCALE)

CUT_AFTER_Q1_HINT = 400   # y in pts, after (1) hint text ends (~394.7)
CUT_BEFORE_Q2 = 435        # y in pts, before (2) starts (~438.5)
CUT_AFTER_Q2_TEXT = 470    # y in pts, after (2) text ends (~466.7)
CUT_BEFORE_Q3 = 517        # y in pts, before (3) starts (~520.5)

cut_q1_hint_px = int(CUT_AFTER_Q1_HINT * SCALE)
cut_before_q2_px = int(CUT_BEFORE_Q2 * SCALE)
cut_after_q2_px = int(CUT_AFTER_Q2_TEXT * SCALE)
cut_before_q3_px = int(CUT_BEFORE_Q3 * SCALE)

doc = fitz.open(INPUT_PDF)
page = doc[0]

mat = fitz.Matrix(SCALE, SCALE)
pix = page.get_pixmap(matrix=mat, alpha=False)

img_data = pix.tobytes("png")
img = Image.open(io.BytesIO(img_data))
w, h = img.size
print(f"Original image: {w}x{h} px")

strip_top = img.crop((0, 0, w, cut_q1_hint_px))
strip_q1_gap = img.crop((0, cut_q1_hint_px, w, cut_before_q2_px))
strip_q2_text = img.crop((0, cut_before_q2_px, w, cut_after_q2_px))
strip_q2_gap = img.crop((0, cut_after_q2_px, w, cut_before_q3_px))
strip_bottom = img.crop((0, cut_before_q3_px, w, h))

new_h = h + extra_px_q1 + extra_px_q2
new_img = Image.new("RGB", (w, new_h), (255, 255, 255))

y = 0
new_img.paste(strip_top, (0, y))
y += strip_top.height

new_img.paste(strip_q1_gap, (0, y))
y += strip_q1_gap.height
y += extra_px_q1

new_img.paste(strip_q2_text, (0, y))
y += strip_q2_text.height

new_img.paste(strip_q2_gap, (0, y))
y += strip_q2_gap.height
y += extra_px_q2

new_img.paste(strip_bottom, (0, y))

print(f"New image: {w}x{new_h} px")

buf = io.BytesIO()
new_img.save(buf, format="PNG", dpi=(DPI, DPI))
buf.seek(0)

page_w_pts = w / SCALE
page_h_pts = new_h / SCALE

out_doc = fitz.open()
out_page = out_doc.new_page(width=page_w_pts, height=page_h_pts)
out_page.insert_image(fitz.Rect(0, 0, page_w_pts, page_h_pts), stream=buf.read())
out_doc.save(OUTPUT_PDF, deflate=True)
out_doc.close()
doc.close()

print(f"Saved to {OUTPUT_PDF}")
print(f"Page size: {page_w_pts:.1f} x {page_h_pts:.1f} pts")
