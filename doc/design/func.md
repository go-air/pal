### Memory

Memory locations are modelled as a set of nodes (as in nodes in a graph), called
_mems_.  For example, mems may correspond to global variables, local variables,
the result of calls to 'malloc', function declarations, etc.  Mems also be specific to
control flow and/or call flow context.  However, pal leaves this opaque to the
user.

Sets of mems may or may not support non-constant values for their size.  For non-constant
values which occur in the program under analysis, a special Value type is provided and
detailed below.

Sets of Mem must provide an efficient means to determine if two mems 'm', 'n' may overlap
and whether they are equal.

The application context will generate a set (or sets) of mems.
