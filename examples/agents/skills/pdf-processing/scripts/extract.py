#!/usr/bin/env python3
"""
Extract text from a PDF file.

Usage:
    extract.py <input_pdf> [output_txt]
"""

import sys
import pdfplumber

def extract_text(pdf_path, output_path=None):
    """Extract text from all pages of a PDF."""
    text_content = []
    
    try:
        with pdfplumber.open(pdf_path) as pdf:
            for i, page in enumerate(pdf.pages, 1):
                print(f"Processing page {i}/{len(pdf.pages)}...", file=sys.stderr)
                text = page.extract_text()
                if text:
                    text_content.append(f"--- Page {i} ---\n{text}\n")
        
        result = '\n'.join(text_content)
        
        if output_path:
            with open(output_path, 'w', encoding='utf-8') as f:
                f.write(result)
            print(f"Text extracted to {output_path}", file=sys.stderr)
        else:
            print(result)
            
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print(__doc__, file=sys.stderr)
        sys.exit(1)
    
    input_pdf = sys.argv[1]
    output_txt = sys.argv[2] if len(sys.argv) > 2 else None
    
    extract_text(input_pdf, output_txt)
