package committer

import (
	"context"

	"cloud.google.com/go/spanner"
)

// CommitPlan represents a collection of mutations to be applied atomically.
type CommitPlan struct {
	mutations []*spanner.Mutation
}

// NewPlan creates a new empty CommitPlan.
func NewPlan() *CommitPlan {
	return &CommitPlan{
		mutations: make([]*spanner.Mutation, 0),
	}
}

// Add adds a mutation to the plan.
// If the mutation is nil, it is ignored.
func (p *CommitPlan) Add(m *spanner.Mutation) {
	if m != nil {
		p.mutations = append(p.mutations, m)
	}
}

// AddAll adds multiple mutations to the plan.
func (p *CommitPlan) AddAll(mutations ...*spanner.Mutation) {
	for _, m := range mutations {
		p.Add(m)
	}
}

// Mutations returns all mutations in the plan.
func (p *CommitPlan) Mutations() []*spanner.Mutation {
	return p.mutations
}

// IsEmpty returns true if the plan has no mutations.
func (p *CommitPlan) IsEmpty() bool {
	return len(p.mutations) == 0
}

// Count returns the number of mutations in the plan.
func (p *CommitPlan) Count() int {
	return len(p.mutations)
}

// Committer applies commit plans to Spanner.
type Committer interface {
	Apply(ctx context.Context, plan *CommitPlan) error
}

// SpannerCommitter implements Committer using Spanner client.
type SpannerCommitter struct {
	client *spanner.Client
}

// NewSpannerCommitter creates a new SpannerCommitter.
func NewSpannerCommitter(client *spanner.Client) *SpannerCommitter {
	return &SpannerCommitter{client: client}
}

// Apply applies all mutations in the plan atomically.
func (c *SpannerCommitter) Apply(ctx context.Context, plan *CommitPlan) error {
	if plan.IsEmpty() {
		return nil
	}

	_, err := c.client.Apply(ctx, plan.Mutations())
	return err
}

// ApplyWithTransaction applies mutations within a read-write transaction.
// This is useful when you need to read data before writing.
func (c *SpannerCommitter) ApplyWithTransaction(
	ctx context.Context,
	fn func(ctx context.Context, txn *spanner.ReadWriteTransaction) (*CommitPlan, error),
) error {
	_, err := c.client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		plan, err := fn(ctx, txn)
		if err != nil {
			return err
		}

		if plan.IsEmpty() {
			return nil
		}

		return txn.BufferWrite(plan.Mutations())
	})

	return err
}
