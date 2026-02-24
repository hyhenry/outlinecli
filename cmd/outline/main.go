package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/henry/outlinecli/internal/api"
	"github.com/henry/outlinecli/internal/config"
	"github.com/henry/outlinecli/internal/credentials"
	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	client *api.Client
)

// requireAuth is used as PersistentPreRunE on commands that need a valid API key.
func requireAuth(cmd *cobra.Command, args []string) error {
	var err error
	cfg, err = config.Load()
	if err != nil {
		return err
	}
	client = api.New(cfg.APIKey, cfg.BaseURL)
	return nil
}

func main() {
	root := &cobra.Command{
		Use:   "outline",
		Short: "Manage notes and documents in Outline",
	}

	root.AddCommand(authCmd(), docCmd(), collectionCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// ─── Output helpers ───────────────────────────────────────────────────────────

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printDocument(doc *api.Document) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID:\t%s\n", doc.ID)
	fmt.Fprintf(w, "Title:\t%s\n", doc.Title)
	fmt.Fprintf(w, "Collection:\t%s\n", doc.CollectionID)
	if doc.ParentDocumentID != "" {
		fmt.Fprintf(w, "Parent:\t%s\n", doc.ParentDocumentID)
	}
	fmt.Fprintf(w, "Created:\t%s\n", doc.CreatedAt)
	fmt.Fprintf(w, "Updated:\t%s\n", doc.UpdatedAt)
	if doc.ArchivedAt != "" {
		fmt.Fprintf(w, "Archived:\t%s\n", doc.ArchivedAt)
	}
	if doc.DeletedAt != "" {
		fmt.Fprintf(w, "Deleted:\t%s\n", doc.DeletedAt)
	}
	w.Flush()
	if doc.Text != "" {
		fmt.Println()
		fmt.Println(doc.Text)
	}
}

func printDocumentList(docs []api.Document) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tUPDATED")
	for _, d := range docs {
		updated := d.UpdatedAt
		if len(updated) > 10 {
			updated = updated[:10]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", d.ID, d.Title, updated)
	}
	w.Flush()
}

func readBodyOrFile(body, file string) (string, error) {
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file %q: %w", file, err)
		}
		return string(data), nil
	}
	return body, nil
}

// ─── auth command ─────────────────────────────────────────────────────────────

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication credentials",
	}
	cmd.AddCommand(authCredentialsCmd(), authStatusCmd(), authRemoveCmd())
	return cmd
}

func authCredentialsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "credentials <env-file>",
		Short: "Load API key from an env file and save credentials",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := credentials.ParseEnvFile(args[0])
			if err != nil {
				return err
			}
			if err := credentials.Save(creds); err != nil {
				return err
			}
			masked := creds.APIKey
			if len(masked) > 12 {
				masked = masked[:10] + "…" + masked[len(masked)-4:]
			}
			fmt.Printf("Credentials saved (%s)\n", masked)
			fmt.Printf("Stored at: %s\n", credentials.Path())
			return nil
		},
	}
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := credentials.Load()
			if err != nil {
				return err
			}

			envKey := os.Getenv("OUTLINE_API_KEY")
			envURL := os.Getenv("OUTLINE_URL")

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if envKey != "" {
				masked := envKey
				if len(masked) > 12 {
					masked = masked[:10] + "…" + masked[len(masked)-4:]
				}
				fmt.Fprintf(w, "API Key (env):\t%s\n", masked)
			} else if creds != nil && creds.APIKey != "" {
				masked := creds.APIKey
				if len(masked) > 12 {
					masked = masked[:10] + "…" + masked[len(masked)-4:]
				}
				fmt.Fprintf(w, "API Key (file):\t%s\n", masked)
			} else {
				fmt.Fprintln(w, "API Key:\tnot set")
			}

			baseURL := envURL
			if baseURL == "" && creds != nil && creds.BaseURL != "" {
				baseURL = creds.BaseURL
			}
			if baseURL == "" {
				baseURL = "https://app.getoutline.com/api (default)"
			}
			fmt.Fprintf(w, "URL:\t%s\n", baseURL)
			fmt.Fprintf(w, "Credentials file:\t%s\n", credentials.Path())
			w.Flush()
			return nil
		},
	}
}

func authRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := credentials.Remove(); err != nil {
				return err
			}
			fmt.Println("Credentials removed.")
			return nil
		},
	}
}

// ─── doc command ─────────────────────────────────────────────────────────────

func docCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "doc",
		Short:             "Manage documents (notes)",
		PersistentPreRunE: requireAuth,
	}
	cmd.AddCommand(
		docAddCmd(),
		docGetCmd(),
		docEditCmd(),
		docDeleteCmd(),
		docListCmd(),
		docSearchCmd(),
		docArchiveCmd(),
		docRestoreCmd(),
	)
	return cmd
}

func docAddCmd() *cobra.Command {
	var (
		title      string
		body       string
		file       string
		collection string
		parent     string
		draft      bool
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a new document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			text, err := readBodyOrFile(body, file)
			if err != nil {
				return err
			}
			doc, err := client.CreateDocument(api.CreateDocumentParams{
				Title:            title,
				Text:             text,
				CollectionID:     collection,
				ParentDocumentID: parent,
				Publish:          !draft,
			})
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(doc)
			}
			fmt.Printf("Created document %s\n", doc.ID)
			printDocument(doc)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Document title (required)")
	cmd.Flags().StringVar(&body, "body", "", "Document content (markdown)")
	cmd.Flags().StringVar(&file, "file", "", "Read content from file path")
	cmd.Flags().StringVar(&collection, "collection", "", "Collection ID")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent document ID")
	cmd.Flags().BoolVar(&draft, "draft", false, "Save as draft instead of publishing")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func docGetCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a document by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := client.GetDocument(args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(doc)
			}
			printDocument(doc)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func docEditCmd() *cobra.Command {
	var (
		title  string
		body   string
		file   string
		asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit a document's title or content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text, err := readBodyOrFile(body, file)
			if err != nil {
				return err
			}
			params := api.UpdateDocumentParams{ID: args[0]}
			if title != "" {
				params.Title = title
			}
			if text != "" {
				params.Text = text
			}
			doc, err := client.UpdateDocument(params)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(doc)
			}
			fmt.Printf("Updated document %s\n", doc.ID)
			printDocument(doc)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&body, "body", "", "New content (markdown)")
	cmd.Flags().StringVar(&file, "file", "", "Read new content from file path")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func docDeleteCmd() *cobra.Command {
	var permanent bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a document (moves to trash by default)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := client.DeleteDocument(args[0], permanent); err != nil {
				return err
			}
			if permanent {
				fmt.Printf("Permanently deleted document %s\n", args[0])
			} else {
				fmt.Printf("Moved document %s to trash\n", args[0])
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&permanent, "permanent", false, "Permanently delete (skips trash)")
	return cmd
}

func docListCmd() *cobra.Command {
	var (
		collection string
		status     string
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List documents",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate status
			validStatuses := map[string]bool{"": true, "draft": true, "archived": true, "published": true}
			if !validStatuses[status] {
				return fmt.Errorf("--status must be one of: draft, archived, published")
			}
			docs, pagination, err := client.ListDocuments(api.ListDocumentsParams{
				CollectionID: collection,
				Status:       status,
			})
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(map[string]any{"data": docs, "pagination": pagination})
			}
			if len(docs) == 0 {
				fmt.Println("No documents found.")
				return nil
			}
			printDocumentList(docs)
			if pagination != nil {
				fmt.Printf("\n%d of %d documents\n", len(docs), pagination.Total)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&collection, "collection", "", "Filter by collection ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: draft, archived, published")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func docSearchCmd() *cobra.Command {
	var (
		collection string
		asJSON     bool
	)
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across documents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := client.SearchDocuments(args[0], collection)
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(results)
			}
			if len(results) == 0 {
				fmt.Println("No results found.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tCONTEXT")
			for _, r := range results {
				ctx := strings.ReplaceAll(r.Context, "\n", " ")
				if len(ctx) > 60 {
					ctx = ctx[:60] + "…"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Document.ID, r.Document.Title, ctx)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().StringVar(&collection, "collection", "", "Limit search to collection ID")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func docArchiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := client.ArchiveDocument(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Archived document %s (%s)\n", doc.ID, doc.Title)
			return nil
		},
	}
}

func docRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <id>",
		Short: "Restore an archived or deleted document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			doc, err := client.RestoreDocument(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Restored document %s (%s)\n", doc.ID, doc.Title)
			return nil
		},
	}
}

// ─── collection command ───────────────────────────────────────────────────────

func collectionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "collection",
		Short:             "Manage collections",
		PersistentPreRunE: requireAuth,
	}
	cmd.AddCommand(collectionListCmd(), collectionGetCmd())
	return cmd
}

func collectionListCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			cols, err := client.ListCollections()
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(cols)
			}
			if len(cols) == 0 {
				fmt.Println("No collections found.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME")
			for _, c := range cols {
				fmt.Fprintf(w, "%s\t%s\n", c.ID, c.Name)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func collectionGetCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a collection by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, err := client.GetCollection(args[0])
			if err != nil {
				return err
			}
			if asJSON {
				return printJSON(col)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID:\t%s\n", col.ID)
			fmt.Fprintf(w, "Name:\t%s\n", col.Name)
			if col.Description != "" {
				fmt.Fprintf(w, "Description:\t%s\n", col.Description)
			}
			fmt.Fprintf(w, "Created:\t%s\n", col.CreatedAt)
			w.Flush()
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}
