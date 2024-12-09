export async function postForm(target, data, token) {
  return fetch(target, {
    method: "POST",
    headers: {
      Authorization: "Bearer " + token,
      "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
    },
    body: Object.keys(data)
      .map((key) => {
        return encodeURIComponent(key) + "=" + encodeURIComponent(data[key]);
      })
      .join("&"),
  })
    .catch((networkError) => {
      console.log("network error:", networkError);
      throw new Error("disconnected from the server");
    })
    .then(async (res) => {
      if (!res.ok) throw new Error(await res.text());
      return res;
    });
}

window.postForm = postForm;
