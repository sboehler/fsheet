# fsheet - create sheet music for flowkey

fsheet is a small utility to generate sheet music in PDF for songs 
in the amazing [flowkey](https://www.flowkey.com) piano learning app, 
for offline playing.

Usage:

1) Log into flowkey in your browser and open a song to generate a PDF for.
2) Save the page (Ctrl-S) to your local disk, e.g. as `flowkey.html`.
3) From the command line, run `fsheet path/to/flowkey.html` (Linux / macOS) or
   `fsheet.exe path\to\flowkey.html` (Windows), which will generate a pdf in your current directory.

