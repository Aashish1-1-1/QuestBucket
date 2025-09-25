function setupPillInput(containerId, inputId, itemsArray) {
  const container = document.getElementById(containerId);
  const input = document.getElementById(inputId);

  function render() {
    container.querySelectorAll(".pill").forEach((el) => el.remove());

    itemsArray.forEach((item, index) => {
      const pill = document.createElement("span");
      pill.className =
        "pill bg-gray-200 text-gray-700 px-3 py-1 rounded-full flex items-center gap-1 text-sm";
      pill.innerHTML = `
          ${item}
          <button type="button" class="text-gray-500 hover:text-red-500" onclick="removeItem('${containerId}', ${index})">×</button>
        `;
      container.insertBefore(pill, input);
    });
  }

  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      let value = input.value.trim().replace(/,$/, "");
      if (value && !itemsArray.includes(value)) {
        itemsArray.push(value);
        render();
      }
      input.value = "";
    }
    if (e.key === "Backspace" && input.value === "" && itemsArray.length > 0) {
      itemsArray.pop();
      render();
    }
  });

  window.removeItem = function (cid, index) {
    if (cid === containerId) {
      itemsArray.splice(index, 1);
      render();
    }
  };
}

let tags = [];

document.addEventListener("DOMContentLoaded", () => {
  setupPillInput("tagContainer", "tagInput", tags);
});

const overlay = document.getElementById("overlay");
const submitFormBtn = document.getElementById("submitFormBtn");

const titleInput = document.getElementById("titleInput");
const descriptionInput = document.getElementById("descriptionInput");
const noteInput = document.getElementById("noteInput");
const openFormBtn = document.getElementById("openFormBtn");

openFormBtn.addEventListener("click", () => {
  overlay.classList.remove("hidden");
});

document.addEventListener("DOMContentLoaded", () => {
  const profileBtn = document.getElementById("profile-btn");
  const popup = document.getElementById("popup");

  // Toggle popup on profile click
  profileBtn.addEventListener("click", (event) => {
    event.stopPropagation(); // prevent click bubbling
    popup.classList.toggle("hidden");
  });

  // Hide popup when clicking outside
  document.addEventListener("click", (event) => {
    if (!profileBtn.contains(event.target) && !popup.contains(event.target)) {
      popup.classList.add("hidden");
    }
  });
});

submitFormBtn.addEventListener("click", () => {
  // Validate fields
  if (!titleInput.value.trim()) {
    alert("Title is required");
    titleInput.focus();
    return;
  }
  if (!descriptionInput.value.trim()) {
    alert("Short description is required");
    descriptionInput.focus();
    return;
  }
  if (tags.length === 0) {
    alert("At least one tag is required");
    document.getElementById("tagInput").focus();
    return;
  }

  const noteData = {
    title: titleInput.value.trim(),
    description: descriptionInput.value.trim(),
    note: noteInput.value.trim(),
    tags: tags,
  };

  overlay.classList.add("hidden");

  console.log("JSON Data:", JSON.stringify(noteData, null, 2));
  const response = postDataWithToken("/addquest", noteData);
  console.log(response);
  titleInput.value = "";
  descriptionInput.value = "";
  noteInput.value = "";
  tags.length = 0;
  document.getElementById("tagInput").value = "";
  document.querySelectorAll("#tagContainer .pill").forEach((el) => el.remove());
});

overlay.addEventListener("click", (e) => {
  if (e.target === overlay) {
    overlay.classList.add("hidden");
  }
});
async function postDataWithToken(url = "", data = {}) {
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    showToast();
    location.reload();
    return await response.json();
  } catch (error) {
    console.error("Error:", error);
  }
}

function showToast(message = "Request successful!") {
  const toast = document.getElementById("toast");
  toast.textContent = message;

  // Show
  toast.classList.remove("opacity-0", "pointer-events-none");
  toast.classList.add("opacity-100");

  // Hide after 5s
  setTimeout(() => {
    toast.classList.add("opacity-0", "pointer-events-none");
    toast.classList.remove("opacity-100");
  }, 5000);
}
function deleteItem(id) {
  if (!confirm("Are you sure you want to delete this item?")) return;

  fetch(`/delete/${id}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
    },
    body: new URLSearchParams({ id }),
  })
    .then((res) => {
      if (!res.ok) throw new Error("Network error");
      return res.json();
    })
    .then((data) => {
      console.log("Deleted:", data);
      showToast();
      location.reload();
    })
    .catch((err) => {
      console.error(err);
      alert("Delete failed: " + err.message);
    });
}
