<script>
  // Input data
  let ip = "";
  let host = "homeassistant.local";

  // Feedback
  let feedback = {
    class: "",
    message: "",
  };

  // Output data
  let key;
  let cert;

  $: if (ip || host) {
    feedback.class = ""
    key = ""
    cert = ""
  }

  let submit = () => {
    feedback = {
      class: "info",
      message: "Working...",
    };

    fetch("/api/generate", {
      method: "post",
      body: JSON.stringify({ host, ip }),
    })
      .then((resp) => {
        if (resp.status == 200) {
          feedback = {
            class: "",
            message: "",
          };
          resp.json().then((data) => {
            key = data.key;
            cert = data.cert;
          });
        } else {
          resp.text().then((text) => {
            feedback = {
              class: "danger",
              message: resp.statusText + ": " + text,
            };
          });
        }
      })
      .catch((e) => {
        feedback = {
          class: "danger",
          message: e,
        };
      });
  };

  function download(filename, text) {
    var element = document.createElement("a");
    element.setAttribute(
      "href",
      "data:text/plain;charset=utf-8," + encodeURIComponent(text),
    );
    element.setAttribute("download", filename);

    element.style.display = "none";
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
  }
</script>

<div class="row">
  <div class="col"></div>

  <div class="col-auto mt-4">
    <h1>Certificate generator</h1>

    <form on:submit|preventDefault={submit}>
      <div class="mb-3">
        <label for="ipaddress" class="form-label">IP Address</label>
        <input
          type="text"
          class="form-control"
          id="ipaddress"
          aria-describedby="ipaddressHelp"
          bind:value={ip}
        />
        <div id="ipaddressHelp" class="form-text">
          The IP address of your Home Assistant server e.g. 192.168.5.53
        </div>
      </div>
      <div class="mb-3">
        <label for="hostname" class="form-label">Host name</label>
        <input
          type="text"
          class="form-control"
          id="hostname"
          bind:value={host}
        />
        <div id="hostHelp" class="form-text">
          The host name of your Home Assistant server. The default is
          homeassistant.local
        </div>
      </div>
      <button type="submit" class="btn btn-primary"
        >Generate certificates</button
      >
    </form>

    {#if feedback.class}
      <div class={`alert alert-${feedback.class} mt-4`}>
        {feedback.message}
      </div>
    {/if}

    {#if key && cert}
      <div class="mt-4">
        <h4>
          Private key
          <button
            type="button"
            class="btn btn-primary float-end"
            on:click={() => download("privkey.pem", key)}>Download</button
          >
        </h4>
        <div class="clearfix"></div>

        <textarea class="form-control mt-2" rows=4 disabled>{key}</textarea>
      </div>

      <div class="mt-4">
        <h4>
          Certificate
          <button
            type="button"
            class="btn btn-primary float-end"
            on:click={() => download("fullchain.pem", key)}>Download</button
          >
        </h4>
        <div class="clearfix"></div>
        <textarea class="form-control mt-2" rows=4 disabled>{cert}</textarea>
      </div>
    {/if}
  </div>

  <div class="col"></div>
</div>
