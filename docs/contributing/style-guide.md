---
layout: default
title: Style Guide
parent: Contributing
nav_order: 4
---

# Style Guide

This is a general guide to the code style we aim to abide by. This is by no means an exhaustive list and we're
constantly changing and improving it. This is also a work in progress document.

For our commit messages please see our [Commit Message Guidelines](./commitmsg-guidelines.md).

## Tools

We implement the following tools that help us abide by our style guide and include the configuration for them inside
our repository:
- [golangci-lint](https://github.com/golangci/golangci-lint)
- [yamllint](https://yamllint.readthedocs.io/en/stable/)
- [eslint](https://eslint.org/)
- [prettier](https://prettier.io/)

## Exceptions

This is a style **guide** not a cudgel, there are always going to be exceptions to these guidelines when it makes sense 
not to follow them. One notable exception is the README.md for the repository. The line length of the 
[All Contributors](https://allcontributors.org/) individual sections are longer than 120 characters and it doesn't make
sense to apply the [line length](#line-length) guidelines.

## Specific Guidelines

### Line Length

We aim to keep all files to a maximum line length of 120 characters. This allows for most modern computer systems to
display two files side by side (vertically split). As always, keep in mind you should not restrict your line length
when it doesn't make sense to.

This includes but is not limited to the following file types:
- Go (*.go)
- YAML (*.yml, *.yaml)
- Markdown (*.md)
- JavaScript (*.js)  
- TypeScript (*.ts, *.tsx)

### Error Strings

Error messages should follow the standard go format. This format can be found in the [golang code review comments](https://github.com/golang/go/wiki/CodeReviewComments#error-strings)
however the key points are:

- errors should not start with capital letters (excluding proper nouns, acronyms, or initialism)
- errors should not end with punctuation
- these restrictions do not apply to logging, only the error type itself

### Configuration Documentation

The configuration documentation has a consistent format this section describes it as best as possible. It's recommended
to check additional sections for examples.

#### Layout

The first thing in the configuration documentation should be a description of the area. This is promptly followed by the
configuration heading (h2 / indent 2) which has an example full configuration.

Under the configuration example each option in the configuration needs to be documented with its own heading 
(h3 / indent 3). Immediately following the heading is a div with some stylized icons. 

The body of the section is to contain paragraphs describing the usage and information specific to that value.


**Example Stylized Icons:**

```html
<div markdown="1">
type: string
{: .label .label-config .label-purple } 
default: example
{: .label .label-config .label-blue }
required: no
{: .label .label-config .label-green }
</div>
```

##### type

This section has the type of the value in a semi human readable format. Some example values are `string`, `integer`, 
`boolean`, `list(string)`, `duration`. This is immediately followed by the styles `.label`, `.label-config`, 
`.label-purple`.

##### default

This section has the default of the value if one exists, this section can be completely omitted if there is no default.
This is immediately followed by the styles `.label`, `.label-config`,
`.label-blue`.

##### required

This section has the required status of the value and must be one of `yes`, `no`, or `situational`. Situational means it
depends on other configuration options. If it's situational the situational usage should be documented. This is 
immediately followed by the styles `.label`, `.label-config`, and a traffic lights color label, i.e. if yes `.label-red`, 
if no `.label-green`, or if situational `.label-yellow`.
