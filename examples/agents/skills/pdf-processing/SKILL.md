---
name: pdf-processing
description: Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.
license: MIT
metadata:
  author: example-org
  version: "1.0"
---

# PDF Processing Skill

## When to use this skill

Use this skill when the user needs to:
- Extract text from PDF files
- Extract tables from PDF documents
- Fill PDF forms programmatically
- Merge multiple PDF files into one
- Split PDF files
- Convert PDFs to other formats

## How to extract text

1. Install the required library:
   ```bash
   pip install pdfplumber
   ```

2. Use the following Python code:
   ```python
   import pdfplumber
   
   with pdfplumber.open("document.pdf") as pdf:
       for page in pdf.pages:
           text = page.extract_text()
           print(text)
   ```

## How to extract tables

```python
import pdfplumber

with pdfplumber.open("document.pdf") as pdf:
    for page in pdf.pages:
        tables = page.extract_tables()
        for table in tables:
            # Process table data
            print(table)
```

## How to fill forms

Use PyPDF2 to fill PDF forms:

```python
from PyPDF2 import PdfReader, PdfWriter

reader = PdfReader("form.pdf")
writer = PdfWriter()

# Update form fields
writer.append_pages_from_reader(reader)
writer.update_page_form_field_values(
    writer.pages[0],
    {"field_name": "field_value"}
)

with open("filled_form.pdf", "wb") as output:
    writer.write(output)
```

## How to merge PDFs

```python
from PyPDF2 import PdfMerger

merger = PdfMerger()
merger.append("file1.pdf")
merger.append("file2.pdf")
merger.write("merged.pdf")
merger.close()
```

## Best practices

- Always close file handles properly
- Handle encrypted PDFs with appropriate passwords
- Validate PDF structure before processing
- Use error handling for corrupted files
