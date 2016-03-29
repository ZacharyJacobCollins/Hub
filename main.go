package main

import (
  "net/http"

  //Third party packages
  // "github.com/ZacharyJacobCollins/Wiki/wiki"
  "github.com/zacharyjacobcollins/HubHtml/chat"
  "github.com/zacharyjacobcollins/login"
)

func main() {
  // w := wiki.NewWiki();  w.Run();
  c := chat.NewChat();  c.Run(3);
  http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./html/"))))
  //http.HandleFunc("/login", chat.Login)
  http.HandleFunc("/index", login.indexPageHandler)
  http.HandleFunc("/internal", login.internalPageHandler)
  http.HandleFunc("/login", login.loginHandler).Methods("POST")
  http.HandleFunc("/logout", login.logoutHandler).Methods("POST")

  http.ListenAndServe(":1337", nil)
}
