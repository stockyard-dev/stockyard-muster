package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-muster/internal/server";"github.com/stockyard-dev/stockyard-muster/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8910"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./muster-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("muster: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Muster — deployment tracker\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("muster: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
