package germanarticlebot

import (
	"github.com/DeryabinSergey/germanarticlebot/internal/domain"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("Invoke", domain.Invoke)
}
