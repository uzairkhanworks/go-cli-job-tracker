package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/mergestat/timediff"
	"github.com/spf13/cobra"
)

const dataFile = "jobs.csv"

type Job struct {
	ID        int
	Role      string
	Company   string
	CreatedAT time.Time
	Applied   bool
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "jobs",
		Short: "Job Application Tracker CLI",
	}
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(appliedCmd())
	rootCmd.AddCommand(deleteCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addCmd() *cobra.Command {
	//jobs add "Full Stack Dev" --copmany JPMorgan
	var company string
	cmd := &cobra.Command{
		Use:   "add <role>",
		Short: "Add a new job application",
		Args:  cobra.ExactArgs(1), //positional arguments => no flag arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			if company == "" {
				return errors.New("Company name required")
			}
			jobs, _ := readJobs()
			id := nextID(jobs)
			job := Job{
				ID:        id,
				Role:      args[0],
				Company:   company,
				CreatedAT: time.Now(),
				Applied:   false,
			}
			jobs = append(jobs, job)
			return writeJobs(jobs)
		},
	}
	cmd.Flags().StringVar(&company, "company", "", "Company name")
	return cmd

}

func listCmd() *cobra.Command {
	var showAll bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Listing all the job applications",
		RunE: func(cmd *cobra.Command, args []string) error {
			jobs, err := readJobs()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tROLE\tCOMPANY\tCREATED\tAPPLIED")
			for _, job := range jobs {
				if !showAll && job.Applied {
					continue
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%v\n",
					job.ID,
					job.Role,
					job.Company,
					timediff.TimeDiff(job.CreatedAT),
					job.Applied)
			}
			w.Flush()
			return nil
		},
	}
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all jobs")
	return cmd
}

func appliedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "applied <id>",
		Short: "Applying for a specific job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("invalid job ID")
			}
			jobs, err := readJobs()
			if err != nil {
				return err
			}
			found := false
			for i := range jobs {
				if jobs[i].ID == id {
					jobs[i].Applied = true
					found = true
					break
				}
			}
			if !found {
				return errors.New("Job Not Found")
			}
			return writeJobs(jobs)
		},
	}
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a specific Job Application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.New("Invalid ID provided")
			}
			jobs, err := readJobs()
			if err != nil {
				return err
			}
			var updated []Job
			found := false
			for _, job := range jobs {
				if job.ID == id {
					found = true
					continue
				}
				updated = append(updated, job)
			}
			if !found {
				return errors.New("Job Not Found")
			}
			return writeJobs(updated)
		},
	}
}

func readJobs() ([]Job, error) {
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Job{}, nil
		}
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	var jobs []Job
	for i, row := range records {
		if i == 0 {
			continue
		}
		id, _ := strconv.Atoi(row[0])
		createdAt, _ := time.Parse(time.RFC3339, row[3])
		applied, _ := strconv.ParseBool(row[4])
		jobs = append(jobs, Job{
			ID:        id,
			Role:      row[1],
			Company:   row[2],
			CreatedAT: createdAt,
			Applied:   applied,
		})
	}
	return jobs, nil

}

func writeJobs(jobs []Job) error {
	file, err := os.Create(dataFile)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"ID", "Role", "Company", "CreatedAt", "Applied"})
	for _, job := range jobs {
		writer.Write([]string{
			strconv.Itoa(job.ID),
			job.Role,
			job.Company,
			job.CreatedAT.Format(time.RFC3339),
			strconv.FormatBool(job.Applied),
		})
	}
	return nil
}

func nextID(jobs []Job) int {
	max := 0
	for _, job := range jobs {
		if job.ID > max {
			max = job.ID
		}
	}
	return max + 1
}
