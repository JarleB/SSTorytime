#

OBJ=postgres_testdb search_broad search_by_node search_clusters search_contrawave_quantum search_simple_text search_by_name search_by_node_text search_contrawave_ambiguous search_interactive search_by_name_unicode search_clusters_functions search_contrawave search_noninteractive search_systematic_knowledge_1 search_systematic_knowledge_2 search_boundary_nodes search_coarse_grain search_coarse_grain2 search_coarse_grain_api search_maze search_maze2 search_wardley

all: $(OBJ)

% : %.go ../../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go


clean:
	rm -f $(OBJ)
	rm -f *~

