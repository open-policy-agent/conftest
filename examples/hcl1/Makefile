NAME := gke

COMMAND := terraform
PLAN = $(NAME)-plan.tfplan
SHOW = $(NAME)-show.json
CODE = $(NAME).tf


all: test

plan: $(PLAN)

$(PLAN): $(CODE)
	$(COMMAND) plan -out $(PLAN)

show: $(SHOW)

$(SHOW): plan
	$(COMMAND) show -json $(PLAN) > $(SHOW)

test: show
	cat $(SHOW) | conftest test -

clean:
	@rm -f $(PLAN) $(SHOW)

.PHONY: plan show test all clean
