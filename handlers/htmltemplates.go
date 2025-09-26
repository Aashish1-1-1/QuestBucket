package handlers

var profile = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.Username}} – Questbucket</title>
    <script src="https://cdn.tailwindcss.com"></script>
  </head>
  <body class="bg-gray-50 text-gray-800">
    <!-- Top bar -->
    <header class="bg-white border-b border-gray-200">
      <div
        class="max-w-6xl mx-auto px-4 py-3 flex flex-wrap justify-between items-center gap-4"
      >
        <!-- Left: Logo -->
        <a href="/" class="flex items-center space-x-2">
          <img src="/assets/logo.svg" alt="Questbucket Logo" class="w-8 h-8" />
          <span class="font-semibold text-lg">Questbucket</span>
        </a>

        <!-- Right: Make your page button -->
        <a
          href="/"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition text-sm font-medium"
        >
          Make your page
        </a>
      </div>
    </header>

    <!-- Profile section -->
    <section class="max-w-6xl mx-auto px-4 py-10">
      <div
        class="flex flex-col md:flex-row items-center md:items-start md:space-x-10 text-center md:text-left"
      >
        <!-- Profile picture -->
        <div class="flex-shrink-0 text-center">
          <img
            src="{{.Pfp_url}}"
            alt="Profile"
            class="w-32 h-32 rounded-full border border-gray-200 object-cover mx-auto md:mx-0"
          />
        <!-- Profile info -->
        <div class="mt-6 md:mt-0">
          <h1 class="text-3xl font-bold text-gray-800">{{.Username}}</h1>
        </div>
        </div>

      </div>
    </section>

    <!-- Quests section (like pinned repos) -->
    <section class="max-w-6xl mx-auto px-4 pb-12">
      <h2 class="text-2xl font-semibold mb-6 text-center md:text-left">
        Quests
      </h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {{range .Quests}}
        <div
          class="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-md transition flex flex-col"
        >
          <a
            href="/post/{{.Questsid.String}}"
            class="text-lg font-semibold text-blue-600 hover:underline"
          >
            {{.Queststitle.String}}
          </a>
          <p class="text-gray-500 text-sm mt-2 flex-grow">
            {{.Questdescription.String}}
          </p>
          <div class="flex flex-wrap gap-2 mt-4">
            {{range .Questtag}}
            <span
              class="bg-blue-50 text-blue-700 px-2 py-1 rounded-full text-xs font-medium"
            >
              {{.}}
            </span>
            {{else}}
            <span class="text-gray-400 text-xs">No tags</span>
            {{end}}
          </div>
        </div>
        {{else}}
        <p class="text-gray-500">No quests found.</p>
        {{end}}
      </div>
    </section>
  </body>
</html>
`
var page = `
					<!doctype html>
					<html lang="en">
					<head>
					  <meta charset="utf-8" />
					  <meta name="viewport" content="width=device-width, initial-scale=1" />
					  <title>QuestBucket</title>
					
					  <!-- EasyMDE CSS -->
					  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.css">
					
					  <!-- Tailwind CSS -->
					  <script src="https://cdn.tailwindcss.com"></script>
					
					  <style>
					    body {
					      font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
					      background: #f8fafc;
					      padding: 24px;
					    }
					    .container {
					      max-width: 900px;
					      margin: 0 auto;
					    }
					  </style>
					</head>
					<body>
					  <div class="container">
					    <h1 class="text-3xl font-bold text-center mb-6">QuestBucket</h1>
					
					    <!-- Textarea that EasyMDE will enhance -->
					    <textarea id="md" name="md" rows="10">{{ .Content }}</textarea>
					
					    <!-- Save Button -->
					    <div class="mt-6 text-center">
					      <button id="saveBtn" class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded shadow-md transition">
					        Save
					      </button>
					    </div>
					  </div>
					
					  <!-- Toast -->
					  <div id="toast" class="fixed bottom-5 left-1/2 -translate-x-1/2 bg-green-600 text-white px-4 py-2 rounded shadow-lg opacity-0 pointer-events-none transition-opacity duration-500">
					    Success
					  </div>
					
					  <!-- EasyMDE JS -->
					  <script src="https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.js"></script>
					
					  <script>
					    // Initialize EasyMDE
					    const easyMDE = new EasyMDE({
					      element: document.getElementById('md'),
					      autosave: { enabled: false },
					      spellChecker: false,
					      toolbar: ["bold", "italic", "heading", "|", "quote", "unordered-list", "ordered-list", "|", "link", "image", "|", "preview", "side-by-side", "fullscreen"],
					      autofocus: true,
					    });
					
					    // Toast logic
					    function showToast() {
					      const toast = document.getElementById('toast');
					      toast.classList.remove('opacity-0', 'pointer-events-none');
					      toast.classList.add('opacity-100');
					      setTimeout(() => {
					        toast.classList.add('opacity-0', 'pointer-events-none');
					        toast.classList.remove('opacity-100');
					      }, 5000);
					    }
					
					    // Save button click
					    const saveBtn = document.getElementById('saveBtn');
					    saveBtn.addEventListener('click', () => {
					      const content = easyMDE.value();
					
					      fetch("/edit/post/{{.Id}}", {
					        method: 'POST',
					        headers: {
					          'Content-Type': 'application/x-www-form-urlencoded',
					        },
					        body: new URLSearchParams({ content })
					      })
					      .then(res => {
					        if (!res.ok) throw new Error('Network response was not ok');
					        return res.json();
					      })
					      .then(data => {
					        // Show toast on success
					        showToast();
					      })
					      .catch(err => {
					        console.error('Save failed', err);
					        alert('Save failed: ' + err.message);
					      });
					    });
					
					    // Optional Ctrl+S shortcut
					    window.addEventListener('keydown', function(e) {
					      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
					        e.preventDefault();
					        saveBtn.click();
					      }
					    });
					  </script>
					</body>
					</html>
				`
