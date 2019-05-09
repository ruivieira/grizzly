package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jawher/mow.cli"
	"github.com/jbrukh/bayesian"
	"github.com/olekukonko/tablewriter"
	"github.com/ruivieira/grizzly"
)

func cmdDuplicate(cmd *cli.Cmd) {
	cmd.Action = func() {

		var notes []grizzly.NoteDuplicate
		grizzly.GetDuplicates(&notes)

		total := 0
		for _, note := range notes {
			total = total + note.Count - 1
		}

		if total > 0 {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"title", "duplicates"})

			for _, note := range notes {

				table.Append([]string{note.Title, strconv.Itoa(note.Count - 1)})
			}
			table.SetFooter([]string{"total", strconv.Itoa(total)}) // Add Footer
			table.Render()

		} else {
			println("👍 no duplicates found.")
		}
	}
}

func cmdNaiveBayes(cmd *cli.Cmd) {
	cmd.Spec = "TITLE"
	title := cmd.StringArg("TITLE", "", "Title to auto-suggest")
	cmd.Action = func() {
		var notes []grizzly.NoteTag
		grizzly.GetAllWithTags(&notes)

		// unique tags
		set := make(map[string]bool)
		for _, note := range notes {
			for _, tag := range note.Tags {
				set[tag] = true
			}
		}
		keys := make([]string, 0)
		for k := range set {
			keys = append(keys, k)
		}
		classes := make([]bayesian.Class, 0)
		for _, v := range keys {
			classes = append(classes, bayesian.Class(v))
		}
		classifier := bayesian.NewClassifier(classes...)
		// train the classifier
		for _, note := range notes {
			titleTokens := strings.Split(note.Title, " ")
			for _, tag := range note.Tags {
				classifier.Learn(titleTokens, bayesian.Class(tag))
			}
		}

		scores, _, _ := classifier.LogScores(strings.Split(*title, " "))

		max := scores[0] // assume first value is the smallest
		index := 0

		for i, value := range scores {
			if value > max {
				max = value
				index = i
			}
		}

		fmt.Printf("Suggest tag for: \"%s\":\n", *title)
		fmt.Printf("\t🏷️  %s (score: %f)\n", keys[index], max)

	}
}

func cmdTail(cmd *cli.Cmd) {
	cmd.Spec = "NUMBER"
	number := cmd.IntArg("NUMBER", 10, "Number of entries (default 10)")

	cmd.Action = func() {

		var notes []grizzly.NoteTag
		grizzly.GetTailWithTags(&notes, *number)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"id", "title", "tags"})

		for _, note := range notes {

			table.Append([]string{strconv.Itoa(note.Id), note.Title, strings.Join(note.Tags, ", ")})
		}
		table.Render()

	}
}

func cmdHead(cmd *cli.Cmd) {
	cmd.Spec = "NUMBER"
	number := cmd.IntArg("NUMBER", 10, "Number of entries (default 10)")

	cmd.Action = func() {

		var notes []grizzly.NoteTag
		grizzly.GetHeadWithTags(&notes, *number)

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"id", "title", "tags"})

		for _, note := range notes {

			table.Append([]string{strconv.Itoa(note.Id), note.Title, strings.Join(note.Tags, ", ")})
		}
		table.Render()

	}
}

func cmdMarkedAll(cmd *cli.Cmd) {

	cmd.Action = func() {

		var notes []grizzly.Note
		grizzly.GetAllMarked(&notes)

		r, _ := regexp.Compile("::(.*)::")
		cb, _ := regexp.Compile("```[a-z]*\\n[\\s\\S]*?\\n```")

		for _, note := range notes {
			// replace code blocks
			text := cb.ReplaceAllLiteralString(note.Text, "")
			matches := r.FindAllString(text, -1)
			var tags string
			if note.Tags == nil {
				tags = ""
			} else {
				tags = fmt.Sprintf("🏷️  %s", strings.Join(note.Tags, ", "))
			}
			fmt.Printf("[#%d %s] %s\n", note.Id, note.Title, tags)
			for _, mark := range matches {
				fmt.Printf("\t%s\n", strings.TrimSuffix(strings.TrimPrefix(mark, "::"), "::"))
			}
		}

		//table := tablewriter.NewWriter(os.Stdout)
		//table.SetHeader([]string{"id", "title", "tags"})
		//
		//for _, note := range notes {
		//
		//	table.Append([]string{strconv.Itoa(note.Id), note.Title, strings.Join(note.Tags, ", ")})
		//}
		//table.Render()

	}
}

func cmdUnlinked(cmd *cli.Cmd) {

	cmd.Action = func() {
		reference := grizzly.GetUnlinked()
		for k, v := range reference {
			if len(v) == 0 {
				fmt.Printf("bear://x-callback-url/open-note?id=%s\n", k)
			}
		}
	}
}

func main() {

	// create an app
	app := cli.App("grizzly", "Bear.app extra utilities")

	// Define our command structure for usage like this:
	app.Command("-d --duplicate", "Find duplicate entries", cmdDuplicate)
	app.Command("-m", "Show all marked passages", cmdMarkedAll)
	app.Command("-ts --tag-suggest", "Suggest tag for title", cmdNaiveBayes)
	app.Command("--tail", "Show oldest notes (by id)", cmdTail)
	app.Command("--head", "Show newest notes (by id)", cmdHead)
	app.Command("-u --unlinked", "Show unlinked notes", cmdUnlinked)

	app.Run(os.Args)
}
