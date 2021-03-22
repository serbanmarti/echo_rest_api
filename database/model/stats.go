package model

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"echo_rest_api/internal"
)

type (
	Stats struct {
		ID        primitive.ObjectID `json:"id"`
		Location  string             `json:"location"`
		ChartType string             `json:"chartType"`
		IsInside  bool               `json:"-"`
		Start     time.Time          `json:"start"`
		End       time.Time          `json:"end"`
		Timezone  int                `json:"timezone"`
		AxisX     StatsAxisX         `json:"axisX"`
		Data      []StatsData        `json:"data"`
	}

	StatsAxisX struct {
		ValueFormat  string `json:"valueFormatString"`
		IntervalType string `json:"intervalType"`
	}

	StatsData struct {
		Type         string           `json:"type"`
		Name         string           `json:"name"`
		Legend       bool             `json:"showInLegend"`
		XValueFormat string           `json:"xValueFormatString"`
		Data         []StatsDataPoint `json:"dataPoints"`
	}

	StatsDataPoint struct {
		Label string `json:"label"`
		X     string `json:"x"`
		Y     uint   `json:"y"`
	}

	StatsDataPointRaw struct {
		Year     uint `bson:"year"`
		Month    uint `bson:"month,omitempty"`
		Day      uint `bson:"day,omitempty"`
		Hour     uint `bson:"hour,omitempty"`
		Minute   uint `bson:"minute,omitempty"`
		Second   uint `bson:"second,omitempty"`
		Entered  uint `bson:"entered"`
		Exited   uint `bson:"exited"`
		MaxCount uint `bson:"max_count"`
	}
)

const (
	eventsCollectionName       = "events"
	spaceResultsCollectionName = "space_results"
	year                       = "year"
	month                      = "month"
	day                        = "day"
	hour                       = "hour"
	second                     = "second"
)

// Retrieves gate statistics from the DB
func StatsGetGate(m *mongo.Database, sq *Stats) error {
	// Create a DB connection
	db := m.Collection(eventsCollectionName)

	// Parse the interval type for the date-time grouping
	var group bson.M
	switch sq.AxisX.IntervalType {
	case hour:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
			day:   bson.M{"$dayOfMonth": "$timestamp"},
			hour:  bson.M{"$hour": "$timestamp"},
		}
	case day:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
			day:   bson.M{"$dayOfMonth": "$timestamp"},
		}
	case month:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
		}
	case year:
		group = bson.M{
			year: bson.M{"$year": "$timestamp"},
		}
	}

	// Parse the direction of data
	var direction int8
	switch sq.IsInside {
	case true:
		direction = -1
	case false:
		direction = 1
	}

	// Aggregate all matching stats entries
	cur, err := db.Aggregate(context.TODO(), []bson.M{
		{
			"$match": bson.M{
				"gate_id": sq.ID,
				"timestamp": bson.M{
					"$gte": sq.Start,
					"$lte": sq.End,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": group,
				"entered": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$crossed", direction}},
							"then": 1,
							"else": 0,
						},
					},
				},
				"exited": bson.M{
					"$sum": bson.M{
						"$cond": bson.M{
							"if":   bson.M{"$eq": []interface{}{"$crossed", -direction}},
							"then": 1,
							"else": 0,
						},
					},
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []string{
						"$_id",
						"$$ROOT",
					},
				},
			},
		},
		{
			"$project": bson.M{"_id": 0, "gate_id": 0, "timestamp": 0},
		},
	})
	if err != nil {
		return internal.NewError(internal.ErrDBQuery, err, 1)
	}

	// Instantiate the entered and exited chart entries
	entered := StatsData{
		Type:         sq.ChartType,
		Name:         "Entered",
		Legend:       true,
		XValueFormat: sq.AxisX.ValueFormat,
		Data:         []StatsDataPoint{},
	}
	exited := StatsData{
		Type:         sq.ChartType,
		Name:         "Exited",
		Legend:       true,
		XValueFormat: sq.AxisX.ValueFormat,
		Data:         []StatsDataPoint{},
	}

	// Decode all found information
	for cur.Next(context.TODO()) {
		var row StatsDataPointRaw

		err = cur.Decode(&row)
		if err != nil {
			return internal.NewError(internal.ErrDBDecode, err, 1)
		}

		// Convert the retrieved DB date to a date string
		dateStr, dErr := dateToString(&row, sq.Timezone, sq.AxisX.IntervalType)
		if dErr != nil {
			return dErr
		}

		// Process the row into entered and exited point types
		rowEntered := StatsDataPoint{
			Label: dateStr,
			X:     dateStr,
			Y:     row.Entered,
		}
		rowExited := StatsDataPoint{
			Label: dateStr,
			X:     dateStr,
			Y:     row.Exited,
		}

		entered.Data = append(entered.Data, rowEntered)
		exited.Data = append(exited.Data, rowExited)
	}

	// Sort the data by dates
	sort.Slice(entered.Data, func(i, j int) bool {
		return entered.Data[i].X < entered.Data[j].X
	})
	sort.Slice(exited.Data, func(i, j int) bool {
		return exited.Data[i].X < exited.Data[j].X
	})

	// Add the entered and exited data
	sq.Data = append(sq.Data, entered, exited)

	// Check if any errors occurred
	if err = cur.Err(); err != nil {
		return internal.NewError(internal.ErrDBCursorIterate, err, 1)
	}

	// Close the cursor once finished
	if err = cur.Close(context.TODO()); err != nil {
		return internal.NewError(internal.ErrDBCursorClose, err, 1)
	}

	// Check if any data found
	if len(sq.Data) == 0 {
		return internal.NewError(internal.ErrDBNoData, err, 1)
	}

	return nil
}

// Retrieve space statistics from the DB
func StatsGetSpace(m *mongo.Database, sq *Stats) error {
	// Create a DB connection
	db := m.Collection(spaceResultsCollectionName)

	// Parse the interval type for the date-time grouping
	var group bson.M
	switch sq.AxisX.IntervalType {
	case hour:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
			day:   bson.M{"$dayOfMonth": "$timestamp"},
			hour:  bson.M{"$hour": "$timestamp"},
		}
	case day:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
			day:   bson.M{"$dayOfMonth": "$timestamp"},
		}
	case month:
		group = bson.M{
			year:  bson.M{"$year": "$timestamp"},
			month: bson.M{"$month": "$timestamp"},
		}
	case year:
		group = bson.M{
			year: bson.M{"$year": "$timestamp"},
		}
	}

	// Aggregate all matching stats entries
	cur, err := db.Aggregate(context.TODO(), []bson.M{
		{
			"$match": bson.M{
				"space_id": sq.ID,
				"timestamp": bson.M{
					"$gte": sq.Start,
					"$lte": sq.End,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": group,
				"max_raw": bson.M{
					"$max": "$count",
				},
			},
		},
		{
			"$replaceRoot": bson.M{
				"newRoot": bson.M{
					"$mergeObjects": []string{
						"$_id",
						"$$ROOT",
					},
				},
			},
		},
		{
			"$project": bson.M{
				"year": 1, "month": 1, "day": 1, "hour": 1,
				"max_count": bson.M{"$cond": bson.M{"if": bson.M{"$lt": []interface{}{"$max_raw", 0}}, "then": 0, "else": "$max_raw"}},
			},
		},
		{
			"$project": bson.M{"_id": 0, "space_id": 0, "timestamp": 0, "count": 0, "stale": 0, "max_raw": 0},
		},
	})
	if err != nil {
		return internal.NewError(internal.ErrDBQuery, err, 1)
	}

	// Instantiate the max count chart entry
	maxCount := StatsData{
		Type:         sq.ChartType,
		Name:         "Max count",
		Legend:       true,
		XValueFormat: sq.AxisX.ValueFormat,
		Data:         []StatsDataPoint{},
	}

	// Decode all found information
	for cur.Next(context.TODO()) {
		var row StatsDataPointRaw

		err = cur.Decode(&row)
		if err != nil {
			return internal.NewError(internal.ErrDBDecode, err, 1)
		}

		// Convert the retrieved DB date to a date string
		dateStr, dErr := dateToString(&row, sq.Timezone, sq.AxisX.IntervalType)
		if dErr != nil {
			return dErr
		}

		// Process the row into max count point type
		rowMaxCount := StatsDataPoint{
			Label: dateStr,
			X:     dateStr,
			Y:     row.MaxCount,
		}

		maxCount.Data = append(maxCount.Data, rowMaxCount)
	}

	// Sort the data by dates
	sort.Slice(maxCount.Data, func(i, j int) bool {
		return maxCount.Data[i].X < maxCount.Data[j].X
	})

	// Add the max count data
	sq.Data = append(sq.Data, maxCount)

	// Check if any errors occurred
	if err = cur.Err(); err != nil {
		return internal.NewError(internal.ErrDBCursorIterate, err, 1)
	}

	// Close the cursor once finished
	if err = cur.Close(context.TODO()); err != nil {
		return internal.NewError(internal.ErrDBCursorClose, err, 1)
	}

	// Check if any data found
	if len(sq.Data) == 0 {
		return internal.NewError(internal.ErrDBNoData, err, 1)
	}

	return nil
}

// Convert a raw stats data point date to a date-string
func dateToString(sd *StatsDataPointRaw, tOff int, mt string) (string, error) {
	// Convert date points to time object
	tStr := fmt.Sprintf(
		"%d-%d-%d %d:%d:%d",
		sd.Year, internal.MaxUint(sd.Month, 1), internal.MaxUint(sd.Day, 1), sd.Hour, sd.Minute, sd.Second,
	)
	ct, err := time.Parse("2006-1-2 15:4:5", tStr)
	if err != nil {
		return "", internal.NewError(internal.ErrBETimeConversion, err, 2)
	}

	// Apply the given timezone offset
	ct = ct.Add(time.Hour * time.Duration(tOff))

	// Convert the time to the required format
	var date string
	switch mt {
	case second:
		date = ct.Format("2006-01-02 15:04:05")
	case hour:
		date = ct.Format("2006-01-02 15:00 - 15:") + "59"
	case day:
		date = ct.Format("2006-01-02")
	case month:
		date = ct.Format("2006-01")
	case year:
		date = ct.Format("2006")
	}

	return date, nil
}
