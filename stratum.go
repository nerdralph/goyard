package main
import(
    "fmt"
    "encoding/json"
    "encoding/binary"
    "net"
    "bufio"
//    "os"
//    "strconv"
    "math"
    "time"
    "bytes"
//    "strings"
    "encoding/hex"
//    "hash"
    "github.com/nerdralph/crypto/sha3"
)
//import "git.io/NR/crypto"

const (
    cacheBYTESINIT = 16*1024*1024
    cacheBYTESGROWTH = 128*1024
    cacheROUNDS = 3
    hashBYTES = 64
)

func isPrime(n int32) bool {
    if (n==2)||(n==3) {return true;}
    if n%2 == 0 { return false }
    if n%3 == 0 { return false }
    sqrt := int32(math.Sqrt(float64(n)))
    for i := int32(5); i <= sqrt; i+=6 {
        if n%i == 0 { return false }
        if n%(i+2) == 0 { return false; }
    }
    return true
}

func cacheSize(epoch int) int {
    sz := cacheBYTESINIT + cacheBYTESGROWTH * epoch
    sz -= hashBYTES
    for ; !isPrime(int32(sz / hashBYTES)); sz -= 2 * hashBYTES {}
    return sz
}

func makeCacheFast(epoch int, seed []byte) []byte {
    sz := cacheSize(epoch)
    cache := make([]byte, sz)
	digestStart := sha3.SumK512(seed)
    copy(cache, digestStart[:])
	kf512 := sha3.ReHashK512()
	digest := kf512.Data()
	copy(digest, digestStart[:])

    for pos := hashBYTES; pos < sz; pos += hashBYTES {
		kf512.Hash()
        copy(cache[pos:], digest)
    }

    // Use a low-round version of randmemohash
    rows := sz/hashBYTES
    for i := 0; i < cacheROUNDS; i++ {
        for j := 0; j < rows; j++ {
            var (
                srcOff = ((j - 1 + rows) % rows) * hashBYTES
                dstOff = j * hashBYTES
                xorOff = (binary.LittleEndian.Uint32(cache[dstOff:]) % uint32(rows)) * hashBYTES
            )
            sha3.FastXORWords(digest, cache[srcOff:srcOff+hashBYTES], cache[xorOff:xorOff+hashBYTES])
			//copy(digest, temp)
			kf512.Hash()
            copy(cache[dstOff:], digest)
        }
    }

    return cache
}

func makeCache(epoch int, seed []byte) []byte {
    sz := cacheSize(epoch)
    cache := make([]byte, sz)

    digest := sha3.SumK512(seed)
    copy(cache, digest[:])
    for pos := hashBYTES; pos < sz; pos += hashBYTES {
        digest = sha3.SumK512(cache[pos-hashBYTES:pos])
        copy(cache[pos:], digest[:])
    }
    fmt.Println("Finished cache creation stage 1", time.Now())

    // Use a low-round version of randmemohash
    temp := make([]byte, hashBYTES)
    rows := sz/hashBYTES
    for i := 0; i < cacheROUNDS; i++ {
        for j := 0; j < rows; j++ {
            var (
                srcOff = ((j - 1 + rows) % rows) * hashBYTES
                dstOff = j * hashBYTES
                xorOff = (binary.LittleEndian.Uint32(cache[dstOff:]) % uint32(rows)) * hashBYTES
            )
            sha3.FastXORWords(temp, cache[srcOff:srcOff+hashBYTES], cache[xorOff:xorOff+hashBYTES])
            digest = sha3.SumK512(temp)
            copy(cache[dstOff:], digest[:])
        }
    }

    return cache
}

type jhdr struct {
    Id int32 `json:"id"`
    Jsonrpc string `json:"jsonrpc"`
}
type jbody struct {
    Method string `json:"method"`
    Params []string `json:"params"`
}
type jmsg struct{
    jhdr
    jbody
}

func main(){
	// pool := "us-east1.ethereum.miningpoolhub.com:20536"
	pool := "eth-us-east1.nanopool.org:9999"
    //pool := "us1.ethermine.org:4444"
    addr := "0xeb9310b185455f863f526dab3d245809f6854b4d"
    conn, err := net.Dial("tcp", pool)
    defer conn.Close()
    if err != nil { fmt.Println(err) }
    fmt.Println("Connected")

    params := []string{addr}
    login := jmsg{jhdr{1, "2.0"}, jbody{"eth_submitLogin", params}}
    fmt.Println("msg:", login)
    data, jerr := json.Marshal(login)
    data = append(data, byte('\n'))
    conn.Write(data)
    fmt.Printf("Sent: %s",data)

    //const bufSize = 4096
    //buf := make([]byte, bufSize)
    var buf []byte
    reader := bufio.NewReader(conn)
    // skip json result:true message
    response := jhdr{99, ""}
    for ; response.Id != 0; { 
        buf, _ = reader.ReadBytes('\n')
        jerr = json.Unmarshal(buf, &response)
        if jerr != nil { fmt.Println(jerr) }
    }
/*
    resp := string(buf[:n])
    if strings.Contains(resp, ":true}") {
        fmt.Printf("Skip: %s",buf[:n])
    }
*/
    var rcvd struct{Result []string `json:"result"`}
    jerr = json.Unmarshal(buf, &rcvd)
    if jerr != nil { fmt.Println(jerr) }
    result := rcvd.Result
    fmt.Println("Result: ", result)
    seedHex := result[1]
    fmt.Println("Seed: ", result[1])

    //digest := make([]byte, 32)
    var digest [32]byte
    seed, _ := hex.DecodeString(seedHex[2:]) 
    epoch := 0
    for !bytes.Equal(digest[:], seed) {
        digest = sha3.SumK256(digest[:])
        epoch++
    }
    fmt.Printf("Epoch %d seed: %x\n",epoch, digest)
    fmt.Println("Starting makeCache", time.Now())
    cache := makeCache(epoch, seed)
    fmt.Println("Starting makeCacheFast", time.Now())
    cache = makeCacheFast(epoch, seed)
    fmt.Println(time.Now(), "Cache size: ", len(cache)) 
    fmt.Printf("%x\n",cache[len(cache)-32:])
}
