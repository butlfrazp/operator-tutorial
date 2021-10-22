# Updating the CRD Definition

**Note:** This is not the greatest application of a operator rather just an exposure to what an operator is and how to implement it.

## Getting Started

1) 
    Start by navigating to `songs-operator/api/v1alpha1/songsconfig_types.go`.
1) 
    You'll see some different `struct` objects. The main `struct` of interest is `SongsConfigSpec`. We want to add songs with a few key attributes to the `SongsConfigSpec` struct. To do this we'll add another struct above.

    ```golang
    type Song struct {
        Id     int    `json:"id"`
        Artist string `json:"artist"`
        Title  string `json:"title"`
        Genre  string `json:"genre"`
    }
    ```
1)
    We now need to update `SongsConfigSpec` to expect the array of songs. `SongsConfigSpec` should be updated to the following schema.

    ```golang
    type SongsConfigSpec struct {
        Songs []Song `json:"songs"`
    }
    ```
1)
    Save the file and run `make generate` and `make manifests`.
    `make generate` will envoke `api/v1alpha1/zz_generated.deepcopy.go`.
    `make manifests` will generate or update the CRP manifests.
    The new CRD manifest will be created in `config/crd/bases/`
1)
    Run `kubectl create -f ./config/crd/bases/<crd_manifest_name>.yaml` to register the CRD.