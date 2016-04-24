package actions

import (
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dimorinny/watchtower/container"
)

var (
	letters  = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	waitTime = 10 * time.Second
)

func containerFilter(names []string) container.Filter {
	if len(names) == 0 {
		return watchtowerContainersFilter
	}

	return func(c container.Container) bool {
		for _, name := range names {
			if (name == c.Name()) || (name == c.Name()[1:] || c.IsWatchtower()) {
				return false
			}
		}
		return true
	}
}

// Update looks at the running Docker containers to see if any of the images
// used to start those containers have been updated. If a change is detected in
// any of the images, the associated containers are stopped and restarted with
// the new image.
func Update(client container.Client, names []string, cleanup bool) error {
	log.Info("Checking containers for updated images")

	containers, err := client.ListContainers(containerFilter(names))
	if err != nil {
		return err
	}

	for i, container := range containers {
		stale, err := client.IsContainerStale(container)
		if err != nil {
			return err
		}
		containers[i].Stale = stale
	}

	containers, err = container.SortByDependencies(containers)
	if err != nil {
		return err
	}

	checkDependencies(containers)

	// Stop stale containers in reverse order
	for i := len(containers) - 1; i >= 0; i-- {
		container := containers[i]

		if container.IsWatchtower() {
			continue
		}

		if container.Stale {
			if err := client.StopContainer(container, waitTime); err != nil {
				log.Error(err)
			}
		}
	}

	// Restart stale containers in sorted order
	for _, container := range containers {
		if container.Stale {
			if err := client.StartContainer(container); err != nil {
				log.Error(err)
			}

			if cleanup {
				client.RemoveImage(container)
			}
		}
	}

	client.ClearIds()

	return nil
}

func checkDependencies(containers []container.Container) {
	for _, parent := range containers {
		if parent.Stale {
			for i, container := range containers {
				if container.IsDepensOn(parent) {
					containers[i].Stale = true
				}
			}
		}
	}
}

// func checkDependencies(containers []container.Container) {

// 	for i, parent := range containers {
// 		if parent.Stale {
// 			continue
// 		}

// 	LinkLoop:
// 		for _, linkName := range parent.Deps() {
// 			fmt.Println
// 			for _, child := range containers {
// 				fmt.Println(child.ID())
// 				fmt.Println(linkName)
// 				if child.ID() == linkName && child.Stale {
// 					containers[i].Stale = true
// 					break LinkLoop
// 				}
// 			}
// 		}
// 	}
// }

// func checkDependencies(containers []container.Container) {

// 	for i, parent := range containers {
// 		if parent.Stale {
// 			continue
// 		}

// 	LinkLoop:
// 		for _, linkName := range parent.Links() {
// 			for _, child := range containers {
// 				if child.Name() == linkName && child.Stale {
// 					containers[i].Stale = true
// 					break LinkLoop
// 				}
// 			}
// 		}
// 	}
// }

// Generates a random, 32-character, Docker-compatible container name.
func randName() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
