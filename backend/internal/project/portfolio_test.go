package project

import "testing"

func TestPortfolioGroupsContractValueByTypeWithoutCancelledProjects(t *testing.T) {
	planned := 40
	portfolio := newPortfolio([]PortfolioProjectResponse{
		{Name: "A", Type: "Perumahan", Status: statusOngoing, Progress: 30, Planned: &planned, ContractValue: 100},
		{Name: "B", Type: "Perumahan", Status: statusCompleted, Progress: 100, ContractValue: 50},
		{Name: "C", Type: "Hotel", Status: statusCancelled, ContractValue: 900},
	})

	if len(portfolio.Projects) != 3 {
		t.Fatalf("projects = %d, want all three kept for the list", len(portfolio.Projects))
	}
	if len(portfolio.ByType) != 1 || portfolio.ByType[0].Count != 2 || portfolio.ByType[0].ContractValue != 150 {
		t.Errorf("by type = %+v, want only Perumahan with 2 projects and 150", portfolio.ByType)
	}
}
