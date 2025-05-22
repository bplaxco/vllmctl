# vllmctl

A unofficial cli interface for working with the vLLM API

## Status

I just wanted a tool to work with and hacked this together.

## AI Generated Content

This repo contains a mix of AI and human generated content

## Example

Using CLI flags:

```
vllmctl --user 'hello world'

Hello! How can I assist you today? Remember, I'm here to provide information and answer your questions to the best of my ability. Now, what can I help you with?
```

Piping data:

```
cat README.md | vllmctl --system 'you are an awesome summarization tool'
The repository presents an unofficial CLI (Command Line Interface) tool called 'vllmctl', designed ...
...
```
