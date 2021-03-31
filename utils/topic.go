package utils

import (
	"regexp"

	"github.com/juju/errors"
)

type TopicWithQos struct {
	Topic string
	Qos   int
}

type TopicInfo struct {
	Organization string
	ID           string
	Type         string
}

// Uuid regex
const UUIDRegex = "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"

// this regex should match this topic /a/d/${UUID}/+
const TopicDeviceRegex = "/([0-9a-f]{1,4})/(d)/(" + UUIDRegex + ")/.*"
const TopicThingRegex = "/([0-9a-f]{1,4})/(t)/(" + UUIDRegex + ")/.*"

var reTopicDevice = regexp.MustCompile(TopicDeviceRegex)
var reTopicThing = regexp.MustCompile(TopicThingRegex)

func IsTopicDeviceThingValid(topic string) bool {
	return reTopicThing.MatchString(topic) ||
		reTopicDevice.MatchString(topic)
}
func IsTopicThingValid(topic string) bool {
	return reTopicThing.MatchString(topic)
}
func IsTopicDeviceValid(topic string) bool {
	return reTopicDevice.MatchString(topic)
}
func extractSubTopicValid(topic string) []string {
	r := reTopicDevice.FindStringSubmatch(topic)
	if len(r) == 0 {
		r = reTopicThing.FindStringSubmatch(topic)
	}
	return r
}
func TopicDeviceThingInfoExtractor(topic string) (TopicInfo, error) {
	if IsTopicDeviceThingValid(topic) {
		// Example with this topic "/a/d/b55c39b7-3edd-4fe5-8cee-259875863e66/s/"
		// will generate this array [/a/d/b55c39b7-3edd-4fe5-8cee-259875863e66/ a b55c39b7-3edd-4fe5-8cee-259875863e66]
		// we need the index 1 that is the organization and the index 2 that is the device id
		submatch := extractSubTopicValid(topic)
		if len(submatch) == 4 {
			return TopicInfo{
				Organization: submatch[1],
				ID:           submatch[3],
				Type:         submatch[2],
			}, nil
		}
	}
	return TopicInfo{}, errors.Errorf("Topic bad formatted: %s", topic)
}
