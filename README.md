# Deploy your static website to Digital Ocean Spaces (with the new CDN enabled!)

### Setup

1. Generate a Spaces key in the Digital Ocean control panel. You can do this at https://cloud.digitalocean.com/account/api/tokens
2. Set some stuff,
    SPACES_KEY=A1BCD2EFG3HIJKL4M5NO
    SPACES_SECRET=JjrXabC1defgHijklmn2/o3P4Qr5+StUV+wx6yZAbC
    SPACES_BUCKET=my-website-bucket-name
3. Get this code: `go get -u github.com/bentranter/spacesite`
4. Run this: `go run main.go`

You then need to manually log into the Digital Ocean Spaces control panel, and enable the CDN for your new bucket, since I haven't figure out how to do it programmatically yet. I'll add it eventually though... maybe.
