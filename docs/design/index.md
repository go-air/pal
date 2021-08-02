# pal -- Pointer Analysis Library for Go

## Goals

The goal of pal is to provide a library which can be effectively used for
different kinds of pointer analyses for Go on different intermediate
representations.

### Effective pointer analysis

Pointer analysis (PA) is a core dependency of many static analyses, which have
different needs, such as

1. Providing a sound dynamic call graph. 
This in turn has many applications
	- impact analysis
 	- non-interference analysis
	- almost any interprocedural sound analysis
	- resolving method calls (with more precision)
	- dataflow analysis, eg for security
1. Identifying possible invalid pointer dereferences.
1. Proving that nil pointer dereferences or
out of bounds panics are impossible.
1. Linking traditional numeric, memory-unaware, analysis methods to modern use. 
1. Identifying aliases.

Unfortunately, PA is often or usually done under global program analysis, as
opposed to modularly.  Tools such as Golang's pointer analysis often requires
re-analyzing the standard library.  Larger projects such as Docker or
Kubernetes take even more resources.

In this project, effective pointer analysis means providing a relatively simple
api to meet the most common needs well, and to meet most needs reasonably.

### For different Go IRs

staticcheck [6] has an IR, golang.org/x/tools/go/ssa is a baseline, we are
working on (air)[https://github.com/go-air/air].  We would like pal to be
retargetable to these different IRs.  Perhaps it can be used one day for the Go
gc compiler IR, or other IRs.

However, to be standard, we will start with a golang.org/x/tools/go/ssa
implementation.

## Architecture

Please see [archi](archi.md).

## Functionality

Please see [func](func.md).


## References

[1] Pointer Analysis. Foundations and Trends in Programming Languages Vol. 2, No. 1 (2015) 1–69
2015 Y. Smaragdakis and G. Balatsouras
DOI: 10.1561/2500000014 (https://yanniss.github.io/points-to-tutorial15.pdf)

[2] Infer
Compositional Analysis by means of bi-abduction
Journal of the ACM Volume 58 Issue 6
December 2011 
Article No.: 26pp 1–66https://doi.org/10.1145/2049697.2049700

[3] Andersen, Lars Ole (1994). Program Analysis and Specialization for the C
Programming Language (PDF) (PhD thesis).

[4] Steensgaard, Bjarne (1996). "Points-to analysis in almost linear time" (PDF). POPL '96: Proceedings of the 23rd ACM SIGPLAN-SIGACT symposium on Principles of programming languages. New York, NY, USA: ACM. pp. 32–41. doi:10.1145/237721.237727. ISBN 0-89791-769-3.

[5] @misc{maksimović2021gillian,
      title={Gillian: A Multi-Language Platform for Unified Symbolic Analysis}, 
      author={Petar Maksimović and José Fragoso Santos and Sacha-Élie Ayoun and Philippa Gardner},
      year={2021},
      eprint={2105.14769},
      archivePrefix={arXiv},
      primaryClass(staticcheck.io) 

[7] Zyrianov, Vlas; Newman, Christian D.; Guarnera, Drew T.; Collard, Michael L.; Maletic, Jonathan I. (2019). "srcPtr: A Framework for Implementing Static Pointer Analysis Approaches" (PDF). ICPC '19: Proceedings of the 27th IEEE International Conference on Program Comprehension. Montreal, Canada: IEEE.

[8] golang.org/x/tools/go/pointer
