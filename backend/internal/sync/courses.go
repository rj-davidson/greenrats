package sync

import (
	"context"
	"fmt"
	"time"
)

func (i *Ingester) syncCourses(ctx context.Context) error {
	start := time.Now()
	i.logger.Info("sync started", "type", "courses")

	courses, err := i.ballDontLie.GetCourses(ctx)
	if err != nil {
		SyncErrors.WithLabelValues("courses").Inc()
		SyncRunsTotal.WithLabelValues("courses", "error").Inc()
		return fmt.Errorf("fetch courses: %w", err)
	}

	coursesProcessed := 0
	holesProcessed := 0

	for idx := range courses {
		c := &courses[idx]

		entCourse, err := i.syncService.UpsertCourse(ctx, c)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("upsert course %s: %w", c.Name, err)
			}
			continue
		}
		coursesProcessed++

		holes, err := i.ballDontLie.GetCourseHoles(ctx, c.ID)
		if err != nil {
			if isContextError(err) {
				return fmt.Errorf("fetch holes for %s: %w", c.Name, err)
			}
			continue
		}

		for hIdx := range holes {
			if err := i.syncService.UpsertCourseHole(ctx, entCourse.ID, &holes[hIdx]); err != nil {
				if isContextError(err) {
					return fmt.Errorf("upsert hole %d for %s: %w", holes[hIdx].HoleNumber, c.Name, err)
				}
				continue
			}
			holesProcessed++
		}
	}

	i.recordSync(ctx, "courses")
	duration := time.Since(start)
	SyncDuration.WithLabelValues("courses").Observe(duration.Seconds())
	SyncRecordsProcessed.WithLabelValues("courses", "courses").Add(float64(coursesProcessed))
	SyncRecordsProcessed.WithLabelValues("courses", "holes").Add(float64(holesProcessed))
	SyncRunsTotal.WithLabelValues("courses", "success").Inc()
	LastSyncTimestamp.WithLabelValues("courses").Set(float64(time.Now().Unix()))
	i.logger.Info("sync completed", "type", "courses", "duration", duration, "courses", coursesProcessed, "holes", holesProcessed)
	return nil
}
