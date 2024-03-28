package grpcapi

//This is the gRPC API server
import (
	"context"
	"database/sql"
	"log"
	"net"
	"time"

	pb "github.com/AdekunleDally/mailinglist/proto"
	"google.golang.org/grpc"

	"github.com/AdekunleDally/mailinglist/mdb"
)

type MailServer struct {
	// This is code generated as part of the protoc command. It is a  type
	// embedded in the struct to be able to run it in the code that was generated

	//Once we have this type embedded within any structure that we create, we will be able to
	//launch it with the grpc service.
	pb.UnimplementedMailingListServiceServer
	db *sql.DB
}

func pbEntryTomdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry {
	t := time.Unix(pbEntry.ConfirmedAt, 0)

	return mdb.EmailEntry{
		Id:          pbEntry.Id,
		Email:       pbEntry.Email,
		ConfirmedAt: &t, //result is a time stamp
		OptOut:      pbEntry.OptOut,
	}
}

func mdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry {

	return pb.EmailEntry{
		Id:          mdbEntry.Id,
		Email:       mdbEntry.Email,
		ConfirmedAt: mdbEntry.ConfirmedAt.Unix(), // result is in seconds
		OptOut:      mdbEntry.OptOut,
	}
}

func emailResponse(db *sql.DB, email string) (*pb.EmailResponse, error) {
	entry, err := mdb.GetEmail(db, email)
	if err != nil {
		return &pb.EmailResponse{}, err
	}

	if entry == nil {
		return &pb.EmailResponse{}, nil
	}

	res := mdbEntryToPbEntry(entry)
	return &pb.EmailResponse{EmailEntry: &res}, nil //seding the data/structure back to the client
}

func (s *MailServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("gRPC GetEmail: %v\n", req)
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) GetEmailBatch(ctx context.Context, req *pb.GetEmailBatchRequest) (*pb.GetEmailBatchResponse, error) {
	log.Printf("gRPC GetEmail: %v\n", req)

	params := mdb.GetEmailBatchQueryParams{
		Page:  int(req.Page),
		Count: int(req.Count),
	}

	mdbEntries, err := mdb.GetEmailBatch(s.db, params)
	if err != nil {
		return &pb.GetEmailBatchResponse{}, err
	}

	pbEntries := make([]*pb.EmailEntry, 0, len(mdbEntries))
	for i := 0; i < len(mdbEntries); i++ {
		entry := mdbEntryToPbEntry((&mdbEntries[i]))
		pbEntries = append(pbEntries, &entry)
	}

	return &pb.GetEmailBatchResponse{EmailEntries: pbEntries}, nil
}

func (s *MailServer) CreateEmail(ctx context.Context, req *pb.CreateEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("gRPC CreateEmail: %v\n", req)

	err := mdb.CreateEmail(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

func (s *MailServer) UpdateEmail(ctx context.Context, req *pb.UpdateEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("gRPC UpdateEmail: %v\n", req)
	entry := pbEntryTomdbEntry(req.EmailEntry)
	err := mdb.UpdateEmail(s.db, entry)

	if err != nil {
		return &pb.EmailResponse{}, err
	}

	return emailResponse(s.db, entry.Email)
}
func (s *MailServer) DeleteEmail(ctx context.Context, req *pb.DeleteEmailRequest) (*pb.EmailResponse, error) {
	log.Printf("gRPC DeleteEmail: %v\n", req)

	err := mdb.DeleteEmail(s.db, req.EmailAddr)
	if err != nil {
		return &pb.EmailResponse{}, err
	}
	return emailResponse(s.db, req.EmailAddr)
}

// write a server function to start the server up
func Serve(db *sql.DB, bind string) {
	listener, err := net.Listen("tcp", bind)

	if err != nil {
		log.Fatalf("gRPC server error: failure to bind %v\n", bind)
	}

	grpcServer := grpc.NewServer()

	mailServer := MailServer{db: db}

	//It takes the grpcServer and tells it that we are using the mailServer as the grpc handler

	pb.RegisterMailingListServiceServer(grpcServer, &mailServer)
	log.Printf("gRPC API server listening on %v\n", bind)

	//Starting up the gRPC server
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("gRPC server error: %v\n", err)
	}

}
