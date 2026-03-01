package sync

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rj-davidson/greenrats/ent/course"
	"github.com/rj-davidson/greenrats/ent/coursehole"
	"github.com/rj-davidson/greenrats/internal/external/balldontlie"
)

func TestUpsertCourse_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{
		ID:      123,
		Name:    "Augusta National Golf Club",
		City:    "Augusta",
		State:   strPtr("GA"),
		Country: "USA",
		Par:     72,
		Yardage: strPtr("7475"),
	}

	created, err := svc.UpsertCourse(ctx, bdlCourse)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "Augusta National Golf Club", created.Name)
	assert.Equal(t, 123, *created.BdlID)
	assert.Equal(t, 72, *created.Par)
	assert.Equal(t, 7475, *created.Yardage)
	assert.Equal(t, "Augusta", *created.City)
	assert.Equal(t, "GA", *created.State)
	assert.Equal(t, "USA", *created.Country)
}

func TestUpsertCourse_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{
		ID:   123,
		Name: "Augusta National",
		Par:  72,
	}

	_, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlCourse.Name = "Augusta National Golf Club"
	bdlCourse.Yardage = strPtr("7510")

	updated, err := svc.UpsertCourse(ctx, bdlCourse)

	require.NoError(t, err)
	assert.Equal(t, "Augusta National Golf Club", updated.Name)
	assert.Equal(t, 7510, *updated.Yardage)
}

func TestUpsertCourseHole_Create(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{ID: 1, Name: "Test Course"}
	courseEntity, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlHole := &balldontlie.CourseHole{
		Course:     balldontlie.CourseRef{ID: 1},
		HoleNumber: 12,
		Par:        3,
		Yardage:    155,
	}

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	holes, err := svc.db.CourseHole.Query().
		Where(coursehole.HasCourseWith(course.IDEQ(courseEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, holes, 1)
	assert.Equal(t, 12, holes[0].HoleNumber)
	assert.Equal(t, 3, holes[0].Par)
	assert.Equal(t, 155, *holes[0].Yardage)
}

func TestUpsertCourseHole_Update(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t)

	bdlCourse := &balldontlie.Course{ID: 1, Name: "Test Course"}
	courseEntity, err := svc.UpsertCourse(ctx, bdlCourse)
	require.NoError(t, err)

	bdlHole := &balldontlie.CourseHole{
		Course:     balldontlie.CourseRef{ID: 1},
		HoleNumber: 1,
		Par:        4,
		Yardage:    445,
	}

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	bdlHole.Par = 5
	bdlHole.Yardage = 520

	err = svc.UpsertCourseHole(ctx, courseEntity.ID, bdlHole)
	require.NoError(t, err)

	holes, err := svc.db.CourseHole.Query().
		Where(coursehole.HasCourseWith(course.IDEQ(courseEntity.ID))).
		All(ctx)

	require.NoError(t, err)
	require.Len(t, holes, 1)
	assert.Equal(t, 5, holes[0].Par)
	assert.Equal(t, 520, *holes[0].Yardage)
}
