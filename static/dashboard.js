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
const openFormBtn = document.getElementById("openFormBtn");
const submitFormBtn = document.getElementById("submitFormBtn");

const titleInput = document.getElementById("titleInput");
const descriptionInput = document.getElementById("descriptionInput");
const noteInput = document.getElementById("noteInput");

openFormBtn.addEventListener("click", () => {
  overlay.classList.remove("hidden");
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

    return await response.json();
  } catch (error) {
    console.error("Error:", error);
  }
}
