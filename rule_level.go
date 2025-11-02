package main

// RuleLevel represents the priority level of rules
type RuleLevel int

const (
	ProjectLevel  RuleLevel = 0 // Most important
	AncestorLevel RuleLevel = 1 // Next most important
	UserLevel     RuleLevel = 2
	SystemLevel   RuleLevel = 3 // Least important
)
