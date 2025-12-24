# PDF Processing Reference Guide

## Supported Libraries

### pdfplumber
- Text extraction
- Table detection and extraction
- Image extraction
- Comprehensive metadata access

### PyPDF2
- Form field manipulation
- PDF merging and splitting
- Encryption and decryption
- Rotation and cropping

## Detailed Examples

### Extract Tables

```python
import pdfplumber

with pdfplumber.open('document.pdf') as pdf:
    for page in pdf.pages:
        tables = page.extract_tables()
        for table in tables:
            # Process table data
            print(table)
```

### Fill PDF Forms

```python
from PyPDF2 import PdfReader, PdfWriter

reader = PdfReader('form.pdf')
writer = PdfWriter()

# Get the first page
page = reader.pages[0]

# Update form fields
writer.add_page(page)
writer.update_page_form_field_values(
    writer.pages[0],
    {"field_name": "value"}
)

# Save the filled form
with open('filled_form.pdf', 'wb') as output_file:
    writer.write(output_file)
```

### Merge PDFs

```python
from PyPDF2 import PdfMerger

merger = PdfMerger()

# Add multiple PDFs
merger.append('file1.pdf')
merger.append('file2.pdf')
merger.append('file3.pdf')

# Write merged PDF
merger.write('merged.pdf')
merger.close()
```

## Performance Considerations

- For large PDFs, process pages incrementally
- Use caching for frequently accessed documents
- Consider parallel processing for batch operations
