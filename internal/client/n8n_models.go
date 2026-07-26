package client

type N8NCredentials struct {
    UUID     string `kv:"VM_UUID"`
    Template string `kv:"Template"`
    State    string `kv:"State"`
    URL      string `kv:"URL"`
}
