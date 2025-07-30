- OpenAI embeddings
- HDBscan for clustering those embeddings
- Maybe also ask GPT-4 for a title for the common embeddings


Should I do this in golang? Not many libraries here. Could experiment with a new language, and the database abstraction makes it possible.

Could just straightout switch to Rust. Or python.

Plan 1:

- [ ] Try implementing in Python
- [ ] Try implementing in Rust
- [ ] Try implementing in Golang
- [ ] Choose the best

---

On the bright side, the database layer does provide a level of separation that allows different services to happen at different levels.

https://hdbscan.readthedocs.io/en/latest/how_hdbscan_works.html

Look into vector database and see what they use for clustering?
https://github.com/milvus-io/milvus
https://github.com/philippgille/chromem-go

https://scikit-learn.org/stable/modules/clustering.html

https://pkg.go.dev/github.com/milvus-io/milvus

---

Plan 2:

- [x] Implement in golang, but keep in mind this limitaitions, and consider having elements in other languages in the future?
- [ ] Get some sample titles
- [ ] Add embeddings
- [ ] Integrate into client
- [ ] 

~~Not viable. Only one hbdscan library in golang, and it is a) nondeterministic, b) unmantained, c) wrong, d) nonexhaustive on points. Reasonably angry.~~
