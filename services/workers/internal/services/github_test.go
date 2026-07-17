package services

import (
	"encoding/base64"
	"testing"

	githubprovider "github.com/briheet/kizuna/workers/internal/providers/github"
	"github.com/stretchr/testify/require"
)

func TestBuildRepositoryGraphIncludesReadmeContent(t *testing.T) {
	payload := GithubJobPayload{Config: GithubJobConfig{Owner: "octocat", Repo: "Hello-World"}}
	repository := &githubprovider.Repository{
		Description: stringPointer("My first repository on GitHub!"),
		FullName:    stringPointer("octocat/Hello-World"),
		HTMLURL:     stringPointer("https://github.com/octocat/Hello-World"),
	}
	readme := &githubprovider.RepositoryContent{
		Name:     stringPointer("README"),
		Path:     stringPointer("README"),
		HTMLURL:  stringPointer("https://github.com/octocat/Hello-World/blob/master/README"),
		Encoding: stringPointer("base64"),
		Content:  stringPointer(base64.StdEncoding.EncodeToString([]byte("Hello World!"))),
	}

	graph, err := buildRepositoryGraph(payload, repository, readme)
	require.NoError(t, err)
	require.Len(t, graph.Nodes, 2)
	require.Len(t, graph.Edges, 1)

	require.Equal(t, "github_repository", graph.Nodes[0].Node.NodeType)
	require.Equal(t, "My first repository on GitHub!", graph.Nodes[0].Chunks[0].Content)
	require.Equal(t, "github_readme", graph.Nodes[1].Node.NodeType)
	require.Equal(t, "README", graph.Nodes[1].Node.Title)
	require.Equal(t, "Hello World!", graph.Nodes[1].Chunks[0].Content)
	require.Equal(t, "https://github.com/octocat/Hello-World/blob/master/README", graph.Nodes[1].Node.SourceLink)
	require.Equal(t, "has_readme", graph.Edges[0].EdgeType)
}

func TestBuildRepositoryGraphAllowsMissingReadme(t *testing.T) {
	payload := GithubJobPayload{Config: GithubJobConfig{Owner: "owner", Repo: "empty"}}
	repository := &githubprovider.Repository{
		FullName: stringPointer("owner/empty"),
		HTMLURL:  stringPointer("https://github.com/owner/empty"),
	}

	graph, err := buildRepositoryGraph(payload, repository, nil)
	require.NoError(t, err)
	require.Len(t, graph.Nodes, 1)
	require.Empty(t, graph.Edges)
}

func stringPointer(value string) *string {
	return &value
}
