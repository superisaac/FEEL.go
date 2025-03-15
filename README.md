# FEEL.go

The interpreter of the FEEL language(Friendly Enough Expression
Language) in go, FEEL is broadly used in DMN and BPMN to provide rule
engine and script support, the FEEL.go module can be imported into
other go projects or used as command line executable as FEEL
interpreter.

## Build
* run `make build` to build feel interpreter bin/feel
* run `make test` to run testing

## Examples
```shell

% bin/feel -c '"hello " + "world"'
"hello world"

% bin/feel -c '(function(a, b) a + b)(5, 8)'
13

% bin/feel -c 'if a > 3 then "larger" else "smaller"' -vars '{a: 5}'
"larger"

# dump AST tree instead of evaluating the script
% bin/feel -c 'if a > 3 then "larger" else "smaller"' -ast
(explist (if (> a 3) "larger"  "smaller"))

% bin/feel -c 'some x in [3, 4, 8, 9] satisfies x % 2 = 0'
4

% bin/feel -c 'every x in [3, 4, 8, 9] satisfies x % 2 = 0'
[
  4,
  8
]
```

for more examples please refer to testing

## Use in golang codes
```golang
import (
  feel "github.com/superisaac/FEEL.go"
)

res, err := feel.EvalString(input)

```