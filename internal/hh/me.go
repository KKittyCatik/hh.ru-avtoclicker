package hh

import (
	"context"
	"fmt"
)

type Me struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (c *Client) GetMe(ctx context.Context) (Me, error) {
	var me Me
	if err := c.DoJSONFast(ctx, "GET", "/me", nil, &me); err != nil {
		return Me{}, fmt.Errorf("get me: %w", err)
	}
	return me, nil
}

type ResumeShort struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (c *Client) GetMyResumes(ctx context.Context) ([]ResumeShort, error) {
	var resp struct {
		Items []ResumeShort `json:"items"`
	}
	if err := c.DoJSONFast(ctx, "GET", "/resumes/mine", nil, &resp); err != nil {
		return nil, fmt.Errorf("get my resumes: %w", err)
	}
	return resp.Items, nil
}
