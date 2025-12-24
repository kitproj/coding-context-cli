---
name: pdf-processing
description: Extract text and tables from PDF files, fill PDF forms, and merge multiple PDFs. Use when working with PDF documents or when the user mentions PDFs, forms, or document extraction.
license: Apache-2.0
metadata:
  author: example-org
  version: "1.0"
---

# PDF Processing

## When to use this skill
Use this skill when the user needs to work with PDF files, including:
- Extracting text or tables from PDF documents
- Filling out PDF forms programmatically
- Merging multiple PDF files into one
- Splitting PDF files into separate documents

## How to extract text
1. Use pdfplumber for text extraction:
   ```python
   import pdfplumber
   with pdfplumber.open('document.pdf') as pdf:
       text = pdf.pages[0].extract_text()
   ```

## How to fill forms
1. Use PyPDF2 to fill form fields:
   ```python
   from PyPDF2 import PdfReader, PdfWriter
   reader = PdfReader('form.pdf')
   writer = PdfWriter()
   # Fill fields here
   ```

## How to merge documents
See [the reference guide](references/REFERENCE.md) for details.
