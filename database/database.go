package database

import (
	"context"
	"log"
	"time"

	"github.com/adityaadpandey/go-gpl/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString = "mongodb://admin:admin123@localhost:27018"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return &DB{
		client: client,
	}
}

func (db *DB) GetJob(id string) *model.JobListing {

	jobCollec := db.client.Database("gpl").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}

	var jobListing model.JobListing
	err := jobCollec.FindOne(ctx, filter).Decode(&jobListing)
	if err != nil {
		log.Fatal(err)
	}
	return &jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollec := db.client.Database("gpl").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var jobListings []*model.JobListing

	cursor, err := jobCollec.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context.TODO(), &jobListings); err != nil {
		panic(err)
	}

	return jobListings
}

func (db *DB) CreateJobListing(jobinfo model.CreateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("gpl").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inserted, err := jobCollec.InsertOne(ctx, bson.M{
		"title":       jobinfo.Title,
		"description": jobinfo.Description,
		"company":     jobinfo.Company,
		"url":         jobinfo.URL,
	})
	if err != nil {
		log.Fatal(err)
	}

	insertedId := inserted.InsertedID.(primitive.ObjectID).Hex()

	returnJobListing := model.JobListing{
		ID:          insertedId,
		Title:       jobinfo.Title,
		Description: jobinfo.Description,
		Company:     jobinfo.Company,
		URL:         jobinfo.URL,
	}
	return &returnJobListing
}

func (db *DB) UpdateJobListing(jobId string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollec := db.client.Database("gpl").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updateJobInfo := bson.M{}

	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}

	result := jobCollec.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	var jobListing model.JobListing

	if err := result.Decode(&jobListing); err != nil {
		log.Fatal(err)
	}

	return &jobListing
}

func (db *DB) DeleteJobListing(jobId string) *model.DeleteJobResponse {
	jobCollec := db.client.Database("gpl").Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_id, _ := primitive.ObjectIDFromHex(jobId)
	filter := bson.M{"_id": _id}
	_, err := jobCollec.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}

	return &model.DeleteJobResponse{DeletedJobID: jobId}
}
