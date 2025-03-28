# startup-sensei

startup-sensei is a tool that scrapes show notes and transcripts from select bootstrap startup podcasts, consolidating them into a JSON file and optional chunked files. 
This dataset can then be fed into Large Language Models (LLMs) for analysis, without the hassle of manually compiling data from multiple sources.

## Why Chunking?

If an LLM has a token or size limit, large transcripts can exceed its context window. The chunked dataset allows to submit them in pieces without losing critical content.

## Usage

### Uploading data:

Upload the generated JSON files to the LLM of your choice. For larger files, if your LLM’s context window is limited, use the chunked files and submit them in parts.

### Prompting example:

Once your data is uploaded, you can prompt your LLM for specific insights. For instance, to analyze product pricing trends mentioned in the podcasts, you might use a prompt like:

```
Based on the provided podcast transcripts and show notes, can you summarize the key product pricing strategies. Please include any recurring themes or notable differences.
```

If `podcasts.json` file is present in the root directory the application will desierialize it. This means that addition of new episodes can be done to the dataset without having to scrape the episodes that are already persisted.  

### Running the program:

With time the dataset will need to be updated with new episodes. To run the program use the make command `make run` or the go command `go run .`. 

This will output a single file with all transcripts alongside the individual chunked files.

#### Adjusting the chunk size

To adjust the chunk size or re-run the scraper, edit the `hunkingOption` `size` variable in main.go.






