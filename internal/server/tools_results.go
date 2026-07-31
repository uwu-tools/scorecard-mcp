package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/uwu-tools/scorecard-mcp/internal/model"
	"github.com/uwu-tools/scorecard-mcp/internal/provider"
	"github.com/uwu-tools/scorecard-mcp/internal/scorecardref"
)

// maxCompareRepos bounds the number of repositories compare_repos will process.
const maxCompareRepos = 10

// --- get_repo_score ---

type getRepoScoreInput struct {
	Repo   string `json:"repo" jsonschema:"repository as platform/owner/repo; platform optional (defaults to github.com), gitlab.com also supported"`
	Commit string `json:"commit,omitempty" jsonschema:"optional 40-character hexadecimal commit SHA to pin results to a specific commit"`
}

type checkSummary struct {
	Name   string `json:"name"`
	Score  int    `json:"score" jsonschema:"score from 0-10, or -1 for inconclusive"`
	Reason string `json:"reason,omitempty"`
}

type repoScoreOutput struct {
	Repo        model.RepoRef       `json:"repo"`
	Commit      string              `json:"commit,omitempty" jsonschema:"the resolved commit SHA the results are for"`
	Date        string              `json:"date,omitempty" jsonschema:"when the scan was generated (RFC3339)"`
	Scorecard   model.ScorecardInfo `json:"scorecard"`
	Source      model.Source        `json:"source" jsonschema:"which provider produced the result"`
	Score       float64             `json:"score" jsonschema:"aggregate score 0-10, or -1 for inconclusive"`
	Checks      []checkSummary      `json:"checks"`
	Caveats     []string            `json:"caveats,omitempty"`
	Attribution *model.Attribution  `json:"attribution,omitempty"`
	Complete    bool                `json:"complete" jsonschema:"whether all checks are represented (false for cached results)"`
}

// --- get_check_result ---

type getCheckResultInput struct {
	Repo   string `json:"repo" jsonschema:"repository as platform/owner/repo; platform optional (defaults to github.com)"`
	Check  string `json:"check" jsonschema:"the Scorecard check name, e.g. Branch-Protection (use list_checks to discover names)"`
	Commit string `json:"commit,omitempty" jsonschema:"optional 40-character hexadecimal commit SHA"`
}

type getCheckResultOutput struct {
	Repo        model.RepoRef       `json:"repo"`
	Commit      string              `json:"commit,omitempty"`
	Date        string              `json:"date,omitempty"`
	Scorecard   model.ScorecardInfo `json:"scorecard"`
	Source      model.Source        `json:"source"`
	Check       model.Check         `json:"check"`
	Caveats     []string            `json:"caveats,omitempty"`
	Attribution *model.Attribution  `json:"attribution,omitempty"`
}

// --- compare_repos ---

type compareReposInput struct {
	Repos []string `json:"repos" jsonschema:"repository references to compare, each as platform/owner/repo"`
}

type repoComparison struct {
	Input  string         `json:"input"`
	Repo   *model.RepoRef `json:"repo,omitempty"`
	Score  *float64       `json:"score,omitempty" jsonschema:"aggregate score 0-10, or -1 for inconclusive"`
	Date   string         `json:"date,omitempty"`
	Source model.Source   `json:"source,omitempty"`
	Error  string         `json:"error,omitempty" jsonschema:"set when this repository could not be retrieved"`
}

type compareReposOutput struct {
	Results   []repoComparison `json:"results"`
	Truncated bool             `json:"truncated" jsonschema:"true if the input list exceeded the supported maximum and was truncated"`
	Note      string           `json:"note,omitempty"`
}

// registerResultTools registers the scorecard-results capability tools.
func registerResultTools(s *mcp.Server, p provider.Provider) {
	mcp.AddTool(s, &mcp.Tool{
		Name: "get_repo_score",
		Description: "Return an OpenSSF Scorecard aggregate score and a per-check summary " +
			"(name, score, short reason) for one repository. This is a compact summary; use " +
			"get_check_result for a single check's full details, and explain_check for a " +
			"check's methodology. Data comes from the OpenSSF Scorecard REST API " +
			"(https://api.scorecard.dev).",
		Annotations: readOnlyAnnotations("Get repository Scorecard score", true),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getRepoScoreInput) (*mcp.CallToolResult, repoScoreOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, repoScoreOutput{}, err
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, repoScoreOutput{}, err
		}
		res, err := p.GetResult(ctx, ref, in.Commit)
		if err != nil {
			return nil, repoScoreOutput{}, err
		}
		return nil, toRepoScore(res), nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "get_check_result",
		Description: "Return the full result (score, reason, details) for a single named OpenSSF " +
			"Scorecard check on one repository. Use list_checks to discover valid check names and " +
			"explain_check for a check's methodology and remediation. Data comes from the OpenSSF " +
			"Scorecard REST API.",
		Annotations: readOnlyAnnotations("Get one Scorecard check result", true),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getCheckResultInput) (*mcp.CallToolResult, getCheckResultOutput, error) {
		ref, err := scorecardref.Parse(in.Repo)
		if err != nil {
			return nil, getCheckResultOutput{}, err
		}
		if err := scorecardref.ValidateCommit(in.Commit); err != nil {
			return nil, getCheckResultOutput{}, err
		}
		if in.Check == "" {
			return nil, getCheckResultOutput{}, fmt.Errorf("check is required; use list_checks to discover valid check names")
		}
		res, err := p.GetResult(ctx, ref, in.Commit)
		if err != nil {
			return nil, getCheckResultOutput{}, err
		}
		c, ok := res.FindCheck(in.Check)
		if !ok {
			return nil, getCheckResultOutput{}, fmt.Errorf(
				"check %q not found in results for %s; use list_checks to see valid check names",
				in.Check, ref.String())
		}
		return nil, getCheckResultOutput{
			Repo:        res.Repo,
			Commit:      res.Commit,
			Date:        res.Date,
			Scorecard:   res.Scorecard,
			Source:      res.Source,
			Check:       c,
			Caveats:     res.Caveats,
			Attribution: res.Attribution,
		}, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name: "compare_repos",
		Description: "Return the OpenSSF Scorecard aggregate scores of several repositories for " +
			"comparison. Each repository's score and provenance are returned; a per-repository " +
			"failure is reported inline rather than failing the whole call. If more than the " +
			"supported maximum are supplied, the list is truncated. Data comes from the OpenSSF " +
			"Scorecard REST API.",
		Annotations: readOnlyAnnotations("Compare repository Scorecard scores", true),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in compareReposInput) (*mcp.CallToolResult, compareReposOutput, error) {
		if len(in.Repos) == 0 {
			return nil, compareReposOutput{}, fmt.Errorf("repos is required; provide one or more repository references")
		}

		var out compareReposOutput
		repos := in.Repos
		if len(repos) > maxCompareRepos {
			out.Truncated = true
			out.Note = fmt.Sprintf("only the first %d of %d repositories were compared", maxCompareRepos, len(in.Repos))
			repos = repos[:maxCompareRepos]
		}

		for _, r := range repos {
			entry := repoComparison{Input: r}
			ref, err := scorecardref.Parse(r)
			if err != nil {
				entry.Error = err.Error()
				out.Results = append(out.Results, entry)
				continue
			}
			entry.Repo = &ref
			res, err := p.GetResult(ctx, ref, "")
			if err != nil {
				entry.Error = err.Error()
				out.Results = append(out.Results, entry)
				continue
			}
			score := res.Score
			entry.Score = &score
			entry.Date = res.Date
			entry.Source = res.Source
			out.Results = append(out.Results, entry)
		}
		return nil, out, nil
	})
}

func toRepoScore(res *model.Result) repoScoreOutput {
	summaries := make([]checkSummary, 0, len(res.Checks))
	for _, c := range res.Checks {
		summaries = append(summaries, checkSummary{Name: c.Name, Score: c.Score, Reason: c.Reason})
	}
	return repoScoreOutput{
		Repo:        res.Repo,
		Commit:      res.Commit,
		Date:        res.Date,
		Scorecard:   res.Scorecard,
		Source:      res.Source,
		Score:       res.Score,
		Checks:      summaries,
		Caveats:     res.Caveats,
		Attribution: res.Attribution,
		Complete:    res.Complete,
	}
}
