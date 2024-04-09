# --------------------
# Makefile for Grawkit
# --------------------

# Run `make help` for information on available actions.

# --------------------
# Variable definitions
# --------------------

# Default name for Grawkit executable.
AWK := awk
CMD := $(CURDIR)/grawkit

# Default executables to use.
SHELL := /bin/bash
DIFF  := $(shell which colordiff 2> /dev/null || which diff)

# Test files to execute.
TESTS := $(shell find tests/*)

# Color & style definitions.
BOLD      := \033[1m
UNDERLINE := \033[4m
RED       := \033[31m
GREEN     := \033[32m
BLUE      := \033[36m
RESET     := \033[0m

# ----------------
# Other directives
# ----------------

# Make `help` be the default action when no arguments are passed to `make`.
.DEFAULT_GOAL := help
.PHONY: $(TESTS) test help

# Awk script for extracting Grawkit documentation as Markdown.
define EXTRACT_MARKDOWN
	/^(#|# .*)$$/ {
		if (f==1) {f=0; printf "```\n\n"}
		print substr($$0, 3)
	}
	/^[^#]/ {
		if (f==0) {f=1; printf "\n```awk\n"}
		print
	}
	!NF {
		print
	}
	END {
		if (f==1) {printf "```\n"}
	}
endef
export EXTRACT_MARKDOWN

# ----------------
# Rule definitions
# ----------------

## Build documentation from source file in Markdown format.
doc:
	@awk "$$EXTRACT_MARKDOWN" "$(CMD)"

## Execute test suite, accepts list of specific files to run.
test: $(TESTS)

$(TESTS):
	@printf ">> $(BOLD)Testing file '$@'...$(RESET) "
	@result=$$($(DIFF) -ud \
	           <(awk '/-->/ {f=1;getline;next} f' $@) <($(AWK) -f $(CMD)    \
	           <(awk '/<!--/ {f=1;next} /-->/ {exit} f' $@)) | tail -n +3); \
	if [ -z "$$result" ]; then                                              \
		printf "$(GREEN)OK$(RESET)\n";                                      \
	else                                                                    \
		printf "$(RED)FAIL$(RESET)\n";                                      \
		echo "$$result";                                                    \
		exit 1;                                                             \
	fi                                                                      \

## Show usage information for this Makefile.
help:
	@printf "$(BOLD)Grawkit — The Awksome Git Graph Generator.$(RESET)\n\n"
	@printf "This Makefile contains tasks for processing auxiliary actions, such as\n"
	@printf "generating documentation or running test cases against the test suite.\n\n"
	@printf "$(UNDERLINE)Available Tasks$(RESET)\n\n"
	@awk -F                                                                       \
		':|##' '/^##/ {c=$$2; getline; printf "$(BLUE)%6s$(RESET) %s\n", $$1, c}' \
		$(MAKEFILE_LIST)
	@printf "\n"
