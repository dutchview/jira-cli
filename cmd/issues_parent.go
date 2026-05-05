package cmd

import (
	"fmt"
	"os"

	"github.com/dutchview/jira-cli/internal/api"
)

// IssuesParentCmd manages an issue's parent (in JIRA's UI this is the
// "Epic" of a story under an epic — same field).
type IssuesParentCmd struct {
	Set   IssuesParentSetCmd   `cmd:"" help:"Attach one or more issues to a parent (e.g., add stories to an epic)"`
	Clear IssuesParentClearCmd `cmd:"" help:"Detach one or more issues from their parent"`
}

type IssuesParentSetCmd struct {
	Children []string `arg:"" name:"child" help:"Issue keys to attach to the parent (one or more)"`
	To       string   `required:"" help:"Parent issue key (the 'Epic' for stories under an epic — same field)"`
}

func (c *IssuesParentSetCmd) Run(client *api.Client) error {
	var failed []string
	for _, child := range c.Children {
		err := client.UpdateIssue(child, map[string]interface{}{
			"parent": map[string]string{"key": c.To},
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, "Failed on %s: %v\n", child, err)
			failed = append(failed, child)
			continue
		}
		fmt.Printf("Set parent of %s to %s.\n", child, c.To)
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to set parent on %d issue(s): %v", len(failed), failed)
	}
	return nil
}

type IssuesParentClearCmd struct {
	Children []string `arg:"" name:"child" help:"Issue keys to detach from their parent (one or more)"`
}

func (c *IssuesParentClearCmd) Run(client *api.Client) error {
	var failed []string
	for _, child := range c.Children {
		err := client.UpdateIssue(child, map[string]interface{}{
			"parent": nil,
		})
		if err != nil {
			fmt.Fprintf(os.Stdout, "Failed on %s: %v\n", child, err)
			failed = append(failed, child)
			continue
		}
		fmt.Printf("Cleared parent of %s.\n", child)
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to clear parent on %d issue(s): %v", len(failed), failed)
	}
	return nil
}
