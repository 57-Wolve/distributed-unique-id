package main
import (
        "fmt"
        "runtime"
        "testing"
        "time"
        "github.com/deckarep/golang-set"
        "github.com/stretchr/testify/assert"
)

var startTime int64
var machineID uint64

func getSnowFlake() *SnowFlake {
        var settings Settings
        settings.StartTime = time.Now() // startTime is the current time
        //settings.StartTime = time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC) // starttime is the Jan 01, 2014
        settings.MachineID = mockMachineId
        //var sf *SnowFlake
        sf := NewSnowFlake(settings)
        if sf == nil {
                panic("SnowFlake not created")
        }
        startTime = toSnowFlakeTime(settings.StartTime)
        //ip, _ := lower16BitPrivateIP()
        machineID = uint64(321)
        return sf
}

func mockMachineId()(uint16, error) {
        return 321, nil
}
func nextID(t *testing.T, sf *SnowFlake) uint64 {
        id, err := sf.NextID()
        if err != nil {
                t.Fatal("id not generated")
        }
        return id
}

func TestSnowFlakeOnce(t *testing.T) {
        sf := getSnowFlake()
        sleepTime := uint64(5)
        fmt.Println(sleepTime)
        time.Sleep(time.Duration(sleepTime) * 10 * time.Millisecond)

        id := nextID(t, sf)

        parts := decompose(id)

        actualMSB := parts["msb"]
        if actualMSB != 0 {
                t.Errorf("unexpected msb: %d", actualMSB)
        }

        actualTime := parts["time"]
        if actualTime < sleepTime  {
                t.Errorf("actualTime shold be greater than sleepTime: %d, %d", actualTime, sleepTime)
        }

        actualSequence := parts["sequence"]
        if actualSequence != 0 {
                t.Errorf("unexpected sequence: %d", actualSequence)
        }

        actualMachineID := parts["machine-id"]
        if actualMachineID != machineID {
                t.Errorf("unexpected machine id: %d", actualMachineID)
        }

        fmt.Println("SnowFlake id:", id)
        fmt.Println("decompose:", parts)
}

func TestSnowFlakeConsecutive(t *testing.T) {
        sf := getSnowFlake()
        id1, _ := sf.NextID()
        id2, _ := sf.NextID()
        assert.Equal(t, true, (id1 < id2), "ID Order Mismatch")
}

func TestSnowFlakeRangeConsecutive(t *testing.T) {
        sf := getSnowFlake()
        lower1, upper1, err := sf.NextIDRange()
        if err != nil {
                t.Fatal("id bounds not generated")
        }
        lower2, upper2, err := sf.NextIDRange()
        assert.Equal(t, true, (lower1 < upper1), "Lower Upper mismatch")
        assert.Equal(t, true, (lower2 < upper2), "Lower Upper mismatch")
        assert.Equal(t, true, (lower1 < lower2), "Lower Lower mismatch")
        assert.Equal(t, true, (upper1 < upper2), "Upper Upper mismatch")
}

func currentTime() int64 {
        return toSnowFlakeTime(time.Now())
}

func TestSnowFlakeList(t *testing.T) {
        sf := getSnowFlake()
        idList, err := sf.NextIDs()
        if err != nil {
                t.Fatal("id list not generated")
        }
        lower := idList[0]
        upper := idList[255]
        a := decompose(lower)
        fmt.Println(a)
        assert.Equal(t, 256, len(idList), "Length of ID List should be 256")
        assert.Equal(t, 256, cap(idList), "Capacity of ID List should be 256")
        //assert.Equal(t, uint64(82944), idList[0], "idList start should be 82944")
        //assert.Equal(t, uint64(83199), idList[255], "idList start should be 83199")
        if (a["time"] == 0) {
                fmt.Println("Running asserts for 0")
                assert.Equal(t, uint64(82176), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(82431), upper, "UpperBound mismatch")
        } else if (a["time"] == 1) {
                fmt.Println("Running asserts for 1")
                assert.Equal(t, uint64(16859392), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(16859647), upper, "UpperBound mismatch")
        } else if (a["time"] == 2) {
                fmt.Println("Running asserts for 2")
                assert.Equal(t, uint64(33636608), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(33636863), upper, "UpperBound mismatch")
        }
}

func TestSnowFlakeRange(t *testing.T) {
        sf := getSnowFlake()
        lower, upper, err := sf.NextIDRange()
        if err != nil {
                t.Fatal("id bounds not generated")
        }
        fmt.Println(lower, upper)
        a := decompose(lower)
        b := decompose(upper)
        fmt.Println(a)
        assert.Equal(t, uint64(0), a["sequence"], "Sequence LowerBound mismatch")
        assert.Equal(t, uint64(255), b["sequence"], "Sequence LowerBound mismatch")
        if (a["time"] == 0) {
                fmt.Println("Running asserts for 0")
                assert.Equal(t, uint64(82176), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(82431), upper, "UpperBound mismatch")
        } else if (a["time"] == 1) {
                fmt.Println("Running asserts for 1")
                assert.Equal(t, uint64(16859392), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(16859647), upper, "UpperBound mismatch")
        } else if (a["time"] == 2) {
                fmt.Println("Running asserts for 2")
                assert.Equal(t, uint64(33636608), lower, "LowerBound mismatch")
                assert.Equal(t, uint64(33636863), upper, "UpperBound mismatch")
        }

        assert.Equal(t, uint64(255), (upper - lower), "Upper and Lower Bound Difference Mismatch")
}

func TestSnowFlakeFor10Sec(t *testing.T) {
        sf := getSnowFlake()
        var numID uint32
        var lastID uint64
        var maxSequence uint64

        initial := currentTime()
        current := initial
        for current - initial < 3 {
                id := nextID(t, sf)

                parts := decompose(id)
                numID++

                if id <= lastID {
                        t.Fatal("duplicated id")
                }
                lastID = id

                current = currentTime()

                actualMSB := parts["msb"]
                if actualMSB != 0 {
                        t.Errorf("unexpected msb: %d", actualMSB)
                }

                actualTime := int64(parts["time"])
                overtime := startTime + actualTime - current
                if overtime > 0 {
                        t.Errorf("unexpected overtime: %d", overtime)
                }

                actualSequence := parts["sequence"]
                if maxSequence < actualSequence {
                        maxSequence = actualSequence
                }

                actualMachineID := parts["machine-id"]
                if actualMachineID != machineID {
                        t.Errorf("unexpected machine id: %d", actualMachineID)
                }
                fmt.Printf("id: %d, machineId: %d, msb: %d, time: %d, seq: %d,\n", id, actualMachineID, actualMSB, actualTime, actualSequence)
        }
        fmt.Println("\n")
        if maxSequence != 1<<BitLenSequence-1 {
                t.Errorf("unexpected max sequence: %d", maxSequence)
        }
        fmt.Println("max sequence:", maxSequence)
        fmt.Println("number of id:", numID)
}

func TestSnowFlakeInParallel(t *testing.T) {
        sf := getSnowFlake()
        numCPU := runtime.NumCPU()
        runtime.GOMAXPROCS(numCPU)
        fmt.Println("number of cpu:", numCPU)

        consumer := make(chan uint64)

        const numID = 10000
        generate := func() {
                for i := 0; i < numID; i++ {
                        consumer <- nextID(t, sf)
                }
        }

        const numGenerator = 10
        for i := 0; i < numGenerator; i++ {
                go generate()
        }

        set := mapset.NewSet()
        for i := 0; i < numID*numGenerator; i++ {
                id := <-consumer
                if set.Contains(id) {
                        t.Fatal("duplicated id")
                } else {
                        set.Add(id)
                }
        }
        fmt.Println("number of id:", set.Cardinality())
}

func TestNilSnowFlake(t *testing.T) {
        var startInFuture Settings
        startInFuture.StartTime = time.Now().Add(time.Duration(1) * time.Minute)
        if NewSnowFlake(startInFuture) != nil {
                t.Errorf("SnowFlake starting in the future")
        }

        var noMachineID Settings
        noMachineID.MachineID = func() (uint16, error) {
                return 0, fmt.Errorf("no machine id")
        }
        if NewSnowFlake(noMachineID) != nil {
                t.Errorf("SnowFlake with no machine id")
        }

        var invalidMachineID Settings
        invalidMachineID.CheckMachineID = func(uint16) bool {
                return false
        }
        if NewSnowFlake(invalidMachineID) != nil {
                t.Errorf("SnowFlake with invalid machine id")
        }
}

func pseudoSleep(period time.Duration, sf *SnowFlake) {
        sf.startTime -= int64(period) / snowFlakeTimeUnitScaleFactor
}

func TestNextIDError(t *testing.T) {
        sf := getSnowFlake()
        year := time.Duration(365*24) * time.Hour
        pseudoSleep(time.Duration(174) * year, sf)
        nextID(t, sf)

        pseudoSleep(time.Duration(1) * year, sf)
        _, err := sf.NextID()
        if err == nil {
                t.Errorf("time is not over")
        }
}
