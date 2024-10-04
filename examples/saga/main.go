package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite"
	_ "github.com/mattn/go-sqlite3"
)

// Define our domain models
type FlightReservation struct {
	ID     string
	Flight string
	Date   time.Time
}

type HotelReservation struct {
	ID    string
	Hotel string
	Dates [2]time.Time
}

type CarRental struct {
	ID       string
	Company  string
	CarModel string
	Dates    [2]time.Time
}

type VacationBooking struct {
	Flight FlightReservation
	Hotel  HotelReservation
	Car    CarRental
}

// Define our saga params
type BookVacationParams struct {
	CustomerID  string
	Destination string
	StartDate   time.Time
	EndDate     time.Time
}

// Failure trigger flags
var (
	failFlightBooking bool
	failHotelBooking  bool = true
	failCarRental     bool
)

// Mock service calls with failure triggers
func bookFlight(customerID, destination string, date time.Time) (FlightReservation, error) {
	time.Sleep(time.Second)
	if failFlightBooking {
		return FlightReservation{}, errors.New("failed to book flight")
	}
	return FlightReservation{
		ID:     fmt.Sprintf("FL-%s", customerID),
		Flight: fmt.Sprintf("Flight to %s", destination),
		Date:   date,
	}, nil
}

func bookHotel(customerID, destination string, startDate, endDate time.Time) (HotelReservation, error) {
	time.Sleep(time.Second)
	if failHotelBooking {
		return HotelReservation{}, errors.New("failed to book hotel")
	}
	return HotelReservation{
		ID:    fmt.Sprintf("HT-%s", customerID),
		Hotel: fmt.Sprintf("Hotel in %s", destination),
		Dates: [2]time.Time{startDate, endDate},
	}, nil
}

func rentCar(customerID, destination string, startDate, endDate time.Time) (CarRental, error) {
	time.Sleep(time.Second)
	if failCarRental {
		return CarRental{}, errors.New("failed to rent car")
	}
	return CarRental{
		ID:       fmt.Sprintf("CR-%s", customerID),
		Company:  "Best Car Rentals",
		CarModel: "Economy",
		Dates:    [2]time.Time{startDate, endDate},
	}, nil
}

// Compensation functions
func cancelFlight(reservation FlightReservation) error {
	log.Printf("Attempting to cancel flight reservation: %s", reservation.ID)
	// Simulate API call
	time.Sleep(time.Second)
	log.Printf("Cancelled flight reservation: %s", reservation.ID)
	return nil
}

func cancelHotel(reservation HotelReservation) error {
	log.Printf("Attempting to cancel hotel reservation: %s", reservation.ID)
	// Simulate API call
	time.Sleep(time.Second)
	log.Printf("Cancelled hotel reservation: %s", reservation.ID)
	return nil
}

func cancelCarRental(rental CarRental) error {
	log.Printf("Attempting to cancel car rental: %s", rental.ID)
	// Simulate API call
	time.Sleep(time.Second)
	log.Printf("Cancelled car rental: %s", rental.ID)
	return nil
}

// Define our saga handler
func bookVacationHandler(ctx context.Context, params BookVacationParams) (*VacationBooking, error) {
	tc := tempolite.GetTaskContextFromContext(ctx)
	if tc == nil {
		return nil, fmt.Errorf("TaskContext not found in context")
	}

	booking := &VacationBooking{}

	// Step 1: Book Flight
	err := tc.SagaStep(
		func() error {
			reservation, err := bookFlight(params.CustomerID, params.Destination, params.StartDate)
			if err != nil {
				return err
			}
			booking.Flight = reservation
			return nil
		},
		func() error {
			if booking.Flight.ID != "" {
				return cancelFlight(booking.Flight)
			}
			return nil
		},
	)
	if err != nil {
		log.Printf("Flight booking failed: %v", err)
		return nil, err
	}

	// Step 2: Book Hotel
	err = tc.SagaStep(
		func() error {
			reservation, err := bookHotel(params.CustomerID, params.Destination, params.StartDate, params.EndDate)
			if err != nil {
				return err
			}
			booking.Hotel = reservation
			return nil
		},
		func() error {
			if booking.Hotel.ID != "" {
				return cancelHotel(booking.Hotel)
			}
			return nil
		},
	)
	if err != nil {
		log.Printf("Hotel booking failed: %v", err)
		return nil, err
	}

	// Step 3: Rent Car
	err = tc.SagaStep(
		func() error {
			rental, err := rentCar(params.CustomerID, params.Destination, params.StartDate, params.EndDate)
			if err != nil {
				return err
			}
			booking.Car = rental
			return nil
		},
		func() error {
			if booking.Car.ID != "" {
				return cancelCarRental(booking.Car)
			}
			return nil
		},
	)
	if err != nil {
		log.Printf("Car rental failed: %v", err)
		return nil, err
	}

	// End the saga
	if err := tc.EndSaga(); err != nil {
		log.Printf("Failed to end saga: %v", err)
		return nil, err
	}

	return booking, nil
}

func main() {
	comfy, err := comfylite3.New(comfylite3.WithMemory())
	// comfy, err := comfylite3.New(comfylite3.WithPath("tempolitesaga.db"))
	if err != nil {
		panic(err)
	}

	db := comfylite3.OpenDB(comfy)

	defer db.Close()
	defer comfy.Close()

	// Create Tempolite instance
	tp, err := tempolite.New(context.Background(), db, 3, 3, 3, 3)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	// Register our handler
	tempolite.RegisterSagaHandler(bookVacationHandler)

	// Create a booking request
	bookingParams := BookVacationParams{
		CustomerID:  "CUST-001",
		Destination: "Hawaii",
		StartDate:   time.Now().AddDate(0, 1, 0), // One month from now
		EndDate:     time.Now().AddDate(0, 1, 7), // One week vacation
	}

	// Start the saga
	ctx := context.Background()
	sagaID, err := tp.EnqueueSaga(ctx, "vacation-booking", bookVacationHandler, bookingParams)
	if err != nil {
		log.Fatalf("Failed to start saga: %v", err)
	}

	fmt.Printf("Started saga with ID: %s\n", sagaID)

	go func() {
		// Wait for the saga to complete
		for {
			sagaInfo, err := tp.GetSagaInfo(ctx, "vacation-booking", sagaID)
			if err != nil {
				log.Fatalf("Failed to get saga info: %v", err)
			}

			switch sagaInfo.Status {
			case tempolite.SagaStatusCompleted:
				fmt.Println("Vacation booking completed successfully!")
				return
			case tempolite.SagaStatusFailed:
				fmt.Println("Vacation booking failed.")
				<-time.After(5 * time.Second)
				return
			case tempolite.SagaStatusCancelled:
				fmt.Println("Vacation booking was cancelled.")
				return
			case tempolite.SagaStatusTerminated:
				fmt.Println("Vacation booking was terminated.")
				return
			default:
				fmt.Printf("Saga status: %v, current step: %d\n", sagaInfo.Status, sagaInfo.CurrentStep)
				time.Sleep(time.Second)
			}
		}
	}()

	if err := tp.WaitForCompletion(context.Background()); err != nil {
		log.Fatalf("Failed to wait for completion: %v", err)
	}
}
