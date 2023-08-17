<!-- 
    is-template: true
    template-data-path: data/home.page.main.data.yaml
-->

<!-- https://docs.github.com/ru/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax -->
# GoLang. Usage for system admininstration

## Part 1. Basic using

**GoLang** is a very popular programming language. Syntax is very simple, but right using is hard.

More info on [Golang Dev Web Site](https:/go.dev/)

## Articles

{{range .Articles}}

### {{index . "Title"}}

{{index . "Content"}}

{{end}}
