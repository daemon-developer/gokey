
## Corpus

### What is a Corpus?

A corpus (_plural: corpora_) is a large and structured set of texts used for linguistic analysis. In the context of computer languages or natural language processing, a corpus is a collection of text data used for analysis, training models, or other language-related tasks.

### Corpora Files

In your `user.json` file you can specify a list of corpus files to analyse. Gokey will assemble quartads  (four-part data structures of sequential key presses used for analysis) from all the provided corpora.

### Gathering a Corpus

For a computer language like Go, a corpus would typically consist of a large collection of Go source code files. Your instructions for gathering this corpus are correct and helpful:
Mac/Linux:
```find . -name "*.go" -type f -print0 | xargs -0 cat > golang.txt```

Windows:
```Get-ChildItem -Path . -Filter *.go -Recurse | Get-Content | Set-Content golang.txt```

The above example will concatenate all the .go files into golang.txt

It's important to consider the relative sizes of your corpora, as they directly influence the quartads used for analysis. For example, if you want to lean more towards Go coding patterns rather than general English prose, ensure you have a larger volume of Go code in your corpus. You can compare the file sizes of your different corpora to gauge their relative proportions.

### References

[1] https://github.com/xsznix/keygen