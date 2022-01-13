package serviceapi

import (
	"context"
	"github.com/vointini/vointini/backend/serviceapi/serviceitems"
	"log"
	"time"
)

func rmfromslice(l []int, r int) (nl []int) {
	for _, i := range l {
		if i == r {
			continue
		}

		nl = append(nl, i)
	}

	return nl
}

// generateReoccurringTasks generates new re-occurring task(s)
func (r Service) generateReoccurringTasks(duration time.Duration) {

	go func(dur time.Duration) {
		for range time.Tick(dur) {
			rtl, err := r.storage.ReOccurringTaskList(context.TODO())
			if err != nil {
				log.Println(err)
				continue
			}

			var reoccuringTaskIds []int
			for _, rot := range rtl {
				// Gather all reoccurring task IDs
				reoccuringTaskIds = append(reoccuringTaskIds, rot.Id)
			}

			for _, rot := range rtl {
				tl, err := r.storage.TaskList(context.TODO(), serviceitems.OngoingTasks)
				if err != nil {
					log.Println(err)
					continue
				}

				for _, ti := range tl {
					// Iterate through all open tasks

					if ti.ReoccurringTaskReferenceId == nil {
						// Not autogenerated
						continue
					}

					if *ti.ReoccurringTaskReferenceId != rot.Id {
						continue
					}

					// Remove reoccurring task from the list since it's still open
					reoccuringTaskIds = rmfromslice(reoccuringTaskIds, rot.Id)
				}
			}

			// TODO enumerate tasks which are done and calculate durations

			// Generate new reoccurring task(s)
			for _, rid := range reoccuringTaskIds {
				for _, rot := range rtl {
					if rot.Id != rid {
						continue
					}

					// Generate new task with reference id
					_, err = r.storage.TaskUpdate(context.TODO(),
						serviceitems.TaskUpdate{
							Id:                         -1, // -1 = New
							Title:                      rot.Title,
							ReoccurringTaskReferenceId: &rot.Id,
						})

					if err != nil {
						log.Println(err)
						continue
					}

				}
			}

		} // for each minute
	}(duration)
}

func (r Service) ReOccurringTaskUpdate(ctx context.Context, update serviceitems.ReoccurringTaskUpdate) (newid int, userErrors []UserError, internalError error) {
	if update.Title == `` {
		userErrors = append(userErrors, UserError{
			Field: "title",
			Msg:   "title cannot be empty",
		})
	}

	if len(userErrors) != 0 {
		return update.Id, userErrors, nil
	}

	newid, internalError = r.storage.ReOccurringTaskUpdate(ctx, update)
	if internalError != nil {
		return update.Id, nil, internalError
	}

	return newid, nil, nil
}

func (r Service) ReOccurringTaskList(ctx context.Context) (list []*serviceitems.ReoccurringTask, internalError error) {
	return r.storage.ReOccurringTaskList(ctx)
}
