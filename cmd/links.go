package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/dutchview/jira-cli/internal/api"
)

type LinksCmd struct {
	List   LinksListCmd   `cmd:"" help:"List links on an issue"`
	Add    LinksAddCmd    `cmd:"" help:"Link two issues"`
	Delete LinksDeleteCmd `cmd:"" help:"Delete an issue link"`
	Types  LinksTypesCmd  `cmd:"" help:"List available link types"`
}

// --- List ---

type LinksListCmd struct {
	IssueKey string `arg:"" help:"Issue key (e.g., PROJ-123)"`
	JSON     bool   `short:"j" help:"Output as JSON"`
}

func (c *LinksListCmd) Run(client *api.Client) error {
	links, err := client.GetIssueLinks(c.IssueKey)
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(links)
	}

	if len(links) == 0 {
		fmt.Printf("No links on %s.\n", c.IssueKey)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tDIRECTION\tISSUE\tSUMMARY")
	fmt.Fprintln(w, "--\t----\t---------\t-----\t-------")

	for _, link := range links {
		typeName := "-"
		if link.Type != nil {
			typeName = link.Type.Name
		}

		if link.OutwardIssue != nil {
			summary := truncate(link.OutwardIssue.Fields.Summary, 50)
			direction := "outward"
			if link.Type != nil {
				direction = link.Type.Outward
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", link.ID, typeName, direction, link.OutwardIssue.Key, summary)
		}
		if link.InwardIssue != nil {
			summary := truncate(link.InwardIssue.Fields.Summary, 50)
			direction := "inward"
			if link.Type != nil {
				direction = link.Type.Inward
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", link.ID, typeName, direction, link.InwardIssue.Key, summary)
		}
	}
	w.Flush()

	return nil
}

// --- Add ---

type LinksAddCmd struct {
	InwardIssue  string `arg:"" help:"Inward issue key (e.g., PROJ-123)"`
	OutwardIssue string `arg:"" help:"Outward issue key (e.g., PROJ-456)"`
	Type         string `short:"t" required:"" help:"Link type name (e.g., 'Blocks', '1. Relates')"`
}

func (c *LinksAddCmd) Run(client *api.Client) error {
	if err := client.CreateIssueLink(c.Type, c.InwardIssue, c.OutwardIssue); err != nil {
		return err
	}

	fmt.Printf("Linked %s -> %s (type: %s)\n", c.InwardIssue, c.OutwardIssue, c.Type)
	return nil
}

// --- Delete ---

type LinksDeleteCmd struct {
	LinkID string `arg:"" help:"Link ID (use 'links list' to find IDs)"`
	Force  bool   `short:"f" help:"Skip confirmation"`
}

func (c *LinksDeleteCmd) Run(client *api.Client) error {
	if !c.Force {
		fmt.Printf("Are you sure you want to delete link %s? [y/N] ", c.LinkID)
		var response string
		fmt.Scanln(&response)
		if !strings.EqualFold(response, "y") {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if err := client.DeleteIssueLink(c.LinkID); err != nil {
		return err
	}

	fmt.Printf("Link %s deleted.\n", c.LinkID)
	return nil
}

// --- Types ---

type LinksTypesCmd struct {
	JSON bool `short:"j" help:"Output as JSON"`
}

func (c *LinksTypesCmd) Run(client *api.Client) error {
	types, err := client.GetIssueLinkTypes()
	if err != nil {
		return err
	}

	if c.JSON {
		return printJSON(types)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tINWARD\tOUTWARD")
	fmt.Fprintln(w, "--\t----\t------\t-------")

	for _, t := range types {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID, t.Name, t.Inward, t.Outward)
	}
	w.Flush()

	return nil
}
