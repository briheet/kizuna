package confluence

import (
	"context"
	"io"

	"github.com/briheet/kizuna/workers/internal/config"
	confluencesdk "github.com/ctreminiom/go-atlassian/v2/confluence/v2"
)

type Client struct {
	client *confluencesdk.Client
}

func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	instance, err := confluencesdk.New(nil, cfg.Confluence.Host)
	if err != nil {
		return nil, err
	}

	instance.Auth.SetBasicAuth(cfg.Confluence.Mail, cfg.Confluence.Token)
	instance.Auth.SetUserAgent("curl/7.54.0")

	return &Client{
		client: instance,
	}, nil
}

func (c *Client) ListSpaces(ctx context.Context, req ListSpacesRequest) (*SpaceChunk, *Response, error) {
	return c.client.Space.Bulk(ctx, req.Options, req.Cursor, req.Limit)
}

func (c *Client) GetSpace(ctx context.Context, req SpaceRequest) (*Space, *Response, error) {
	return c.client.Space.Get(ctx, req.SpaceID, req.DescriptionFormat)
}

func (c *Client) ListPages(ctx context.Context, req ListPagesRequest) (*PageChunk, *Response, error) {
	return c.client.Page.Gets(ctx, req.Options, req.Cursor, req.Limit)
}

func (c *Client) GetPage(ctx context.Context, req PageRequest) (*Page, *Response, error) {
	return c.client.Page.Get(ctx, req.PageID, req.Format, req.Draft, req.Version)
}

func (c *Client) ListPagesBySpace(ctx context.Context, req ListPagesBySpaceRequest) (*PageChunk, *Response, error) {
	return c.client.Page.GetsBySpace(ctx, req.SpaceID, req.Cursor, req.Limit)
}

func (c *Client) ListChildPages(ctx context.Context, req ListChildPagesRequest) (*ChildPageChunk, *Response, error) {
	return c.client.Page.GetsByParent(ctx, req.PageID, req.Cursor, req.Limit)
}

func (c *Client) ListPagesByLabel(ctx context.Context, req ListPagesByLabelRequest) (*PageChunk, *Response, error) {
	return c.client.Page.GetsByLabel(ctx, req.LabelID, req.Sort, req.Cursor, req.Limit)
}

func (c *Client) ListAttachments(ctx context.Context, req ListAttachmentsRequest) (*AttachmentPage, *Response, error) {
	return c.client.Attachment.Gets(ctx, req.EntityID, req.EntityType, req.Options, req.Cursor, req.Limit)
}

func (c *Client) GetAttachment(ctx context.Context, req AttachmentRequest) (*Attachment, *Response, error) {
	return c.client.Attachment.Get(ctx, req.AttachmentID, req.VersionID, req.SerializeIDs)
}

func (c *Client) DownloadAttachment(ctx context.Context, attachmentID string) (io.ReadCloser, error) {
	return c.client.Attachment.Download(ctx, attachmentID)
}

func (c *Client) ListCustomContent(ctx context.Context, req ListCustomContentRequest) (*CustomContentPage, *Response, error) {
	return c.client.CustomContent.Gets(ctx, req.Type, req.Options, req.Cursor, req.Limit)
}

func (c *Client) GetCustomContent(ctx context.Context, req CustomContentRequest) (*CustomContent, *Response, error) {
	return c.client.CustomContent.Get(ctx, req.CustomContentID, req.Format, req.VersionID)
}
