# compiler.

For our purposes, we use a simple subset of the regular expressions found in most modern
environments and languages. Most of the below is based on what is found in the third chapter of
_The Dragon Book_.

An _alphabet_ *Σ* is a finite sequence of symbols.

A _string_ over an alphabet is a finite sequence of symbols drawn from that alphabet. The _empty
string_, denoted by *ε*, is the string of length zero.
If *x* and *y* are strings, then the _concatenation_ of *x* and *y*, denoted *xy*, is the string
formed by appending *y* to *x*.

A _language_ L is any countable set of strings over some fixed alphabet.
The _union_ of two languages _L_ and _M_ is the set of strings where each string is in either of the
two languages. The _concatenation_ is the set of strings _st_ where _s_ is in _L_ and _t_ in _M_.
The _(Kleene-)closure_ of _L_, denoted _L*_, is the set of strings you get by concatenating _L_ zero
or more times (with itself); and the _positive closure_ is defined
```
L+ = L* - {ε}.
```

Regular expressions are then defined in two steps. First we define the basis or the basic set of
expressions, and then we add some rules by which we may form more elaboreate expressions.

Each regular expression _r_ denotes a language _L(r)_ over some alphabet _Σ_.

## basis.

1. _ε_ is a regular expression, matching the empty string.
2. If _a_ is a symbol in _Σ_, then _a_ is a regular expression, and
L(a) = {a},
that is, the language with one string, of length one, with _a_ in its one position.

## extensions.

1. _(r)|(s)_ is a regular expression denoting the union of _L(r)_ with _L(s)_
2. _(r)(s)_ is a regular expression denoting the concatenation _L(r)L(s)_
3. _(r)*_ is a regular expression denoting _(L(r))*_
4. _(r)+_ is a regular expression denoting _(L(r))+_
5. _(r)_ is a regular expression denoting _L(r)_.

### dropping parentheses.

We may drop certain pairs of parentheses if we adopt the conventions that

1. The unary operator _*_ has the highest precedence and is left associative
2. Concatenation has second-highest precedence and is left associative.


## our regular expressions.

In our case, our choice of alphabet is the set of digits and letters. This gives us the following
BNF definition (we have used range notation to reduce verbosity; it is assumed that there are no
spaces between subsequent symbols):

```
expr   → expr *
       | expr +
       | seq

seq    → union | union seq

union  → basic | basic '|' basic

basic  → ( expr ) 
       | symbol

symbol → a-Z | A-Z | 0-9
```

