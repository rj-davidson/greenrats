package sync

import (
	"context"
	"time"
)

func (i *Ingester) syncCourses(ctx context.Context) {
	start := time.Now()
	i.logger.Info("sync started", "type", "courses")

	courses, err := i.ballDontLie.GetCourses(ctx)
	if err != nil {
		i.logger.Error("failed to fetch courses", "error", err)
		i.captureJobError("sync_courses", err)
		SyncErrors.WithLabelValues("courses").Inc()
		SyncRunsTotal.WithLabelValues("courses", "error").Inc()
		return
	}

	coursesProcessed := 0
	holesProcessed := 0

	for idx := range courses {
		c := &courses[idx]

		entCourse, err := i.syncService.UpsertCourse(ctx, c)
		if err != nil {
			i.logger.Error("failed to upsert course", "course", c.Name, "error", err)
			i.captureJobError("sync_courses", err)
			continue
		}
		coursesProcessed++

		holes, err := i.ballDontLie.GetCourseHoles(ctx, c.ID)
		if err != nil {
			i.logger.Error("failed to fetch course holes", "course", c.Name, "error", err)
			continue
		}

		for hIdx := range holes {
			if err := i.syncService.UpsertCourseHole(ctx, entCourse.ID, &holes[hIdx]); err != nil {
				i.logger.Error("failed to upsert course hole", "course", c.Name, "hole", holes[hIdx].HoleNumber, "error", err)
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
}
