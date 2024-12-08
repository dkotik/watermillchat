export async function postForm(target, params) {
  return fetch(target, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8",
    },
    body: Object.keys(params)
      .map((key) => {
        return encodeURIComponent(key) + "=" + encodeURIComponent(params[key]);
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
