---
applyTo: '**'
---
Provide project context and coding guidelines that AI should follow when generating code, answering questions, or reviewing changes.

Follow these rules for all responses and code changes:
1. Restrict all code to very simple control flow constructs. Do not use `goto`, `setjmp`, `longjmp`, or direct/indirect recursion.
2. All loops must have a fixed, statically provable upper-bound on iterations, unless intentionally non-terminating; in that case it must be statically provable the loop cannot terminate.
3. Do not use dynamic memory allocation after initialization.
4. No function should exceed about 60 lines of code (single page, one line per statement/declaration).
5. Maintain an average of at least two assertions per function. Assertions must be side-effect free, meaningful, and failures must trigger explicit recovery actions (e.g., return an error).
6. Declare data objects in the smallest possible scope.
7. Always check return values of non-void functions and validate parameters inside each function; explicitly cast to `(void)` only when ignoring is justified.
8. Limit preprocessor use to header inclusion and simple macro definitions; no token pasting, var-args macros, recursive macros, or complex conditional compilation.
9. Restrict pointer use to at most one level of dereference; no hidden dereferences in macros/typedefs and no function pointers.
10. Compile with all warnings at the most pedantic setting and keep builds clean; run at least one static analyzer daily and keep zero warnings.