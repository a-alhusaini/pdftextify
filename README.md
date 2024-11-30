# pdftextify - a basic utility that converts PDFs into .TXT

As I have low vision it is relatively inconvenient to read a PDF file. This little utility uses openai's vision capabilities to extract text from a pdf file

Usage:
`bash convert_book_text.sh my_book.pdf`

Will create a file `my_book.pdf.txt` which has the OCR'd text from the
original document.

Be sure to set your `OPENAI_API_KEY` in your environment to use this. 
