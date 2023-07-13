# FEEL.go
Interpretere of the FEEL language(Friendly Enough Expression Language) in go

## Build
run `make build` to build feel interpreter bin/feel
run `make test` to run testing

## Examples
```
% bin/feel -c '"hello " + "world"'
"hello world"

% bin/feel -c '(function(a, b) a + b)(5, 8)'
13

# dump AST tree instead of evaluating the script
% bin/feel -c 'bind("a", 5); if a > 3 then "larger" else "smaller"' -ast
(explist (call bind ["a", 5]) (if (> a 3) then "larger" else "smaller"))
```

for more examples please refer to testing