package client

type OpenClawCredentials struct {
    UUID     string `kv:"VM_UUID"`
    Template string `kv:"Template"`
    State    string `kv:"State"`
    URL      string `kv:"URL"`
    Token    string `kv:"OPENCLAW_TOKEN"`
}
