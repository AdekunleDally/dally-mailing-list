package main

import (
	"context"
	"log"
	"time"

	pb "github.com/AdekunleDally/mailinglist/proto"
	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func logResponse(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf(" error: %v", err)
	}

	if res.EmailEntry == nil {
		log.Printf(" email not found")
	} else {
		log.Printf(" response: %v", res.EmailEntry)
	}
}

func createEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("create email")
	// cancel a request after 1 second
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

func getEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("get email")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

func getEmailBatch(client pb.MailingListServiceClient, count int, page int) {
	log.Println("get email batch ")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{Count: int32(count), Page: int32(page)})
	if err != nil {
		log.Fatalf(" error: %v", err)
	}
	log.Println("response:")

	// We loop through each item in the response,and then print out the item it is, the total that we get back and the actual response itself,
	for i := 0; i < len(res.EmailEntries); i++ {
		log.Printf("  item [%v of %v]: %s", i+1, len(res.EmailEntries), res.EmailEntries[i])
	}
}

func updateEmail(client pb.MailingListServiceClient, entry pb.EmailEntry) *pb.EmailEntry {
	log.Println("Update email")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: &entry})
	logResponse(res, err)

	return res.EmailEntry
}

func deleteEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("Delete email")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

// Creating the commandline arguments
var args struct {
	GrpcAddr string `args:"env:MAILINGLIST_GRPC_ADDR"`
}

func main() {
	arg.MustParse(&args)

	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8081"
	}

	//connecting to the grpc server
	conn, err := grpc.Dial(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// We then create a new client. When we use the NewMailingListServiceClient() function
	//We are telling the existing grpc connection(i.e conn) that the client is associated with
	// the existing grpc messages we define within our mailing list services allowing us send requests to the server

	client := pb.NewMailingListServiceClient(conn)

	newEmail := createEmail(client, "adeorimi@mideacouture.com")
	newEmail.ConfirmedAt = 10000
	updateEmail(client, *newEmail)
	deleteEmail(client, newEmail.Email)

	getEmailBatch(client, 5, 1)
	getEmailBatch(client, 3, 2)
	getEmailBatch(client, 3, 3)
}
