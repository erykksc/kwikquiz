document.addEventListener('DOMContentLoaded', function() {
    const addQuestinBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');
    addQuestinBtn.addEventListener('click', function () {

        const questionNumber = document.querySelectorAll('.question-item').length + 1;
        const questionHTML = `
        <div class="question-item mb-4 p-4 border rounded-lg">
          <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
          <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" required>
          <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
          <input type="text" name="answer-${questionNumber}-1" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option 1" required>
          <input type="text" name="answer-${questionNumber}-2" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option 2" required>
          <input type="text" name="answer-${questionNumber}-3" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option 3" required>
          <input type="text" name="answer-${questionNumber}-4" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option 4" required>
               <!-- Correct Answer -->
          <div class="mb-4">
            <label for="option-1" class="block text-gray-700 font-semibold mb-2">Correct Answer</label>
            <div class="flex items-center">
              <input type="radio" id="option-1" name="correct-answer-${questionNumber}" value="1" class="mr-2" required>
              <label for="option-1" class="mr-4 text-gray-700">Option 1</label>
              <input type="radio" id="option-2" name="correct-answer-${questionNumber}" value="2" class="mr-2">
              <label for="option-2" class="text-gray-700">Option 2</label>
              <input type="radio" id="option-3" name="correct-answer-${questionNumber}" value="3" class="mr-2">
              <label for="option-3" class="text-gray-700">Option 3</label>
              <input type="radio" id="option-4" name="correct-answer-${questionNumber}" value="4" class="mr-2">
              <label for="option-4" class="text-gray-700">Option 4</label>
            </div>
          </div>
          <button type="button" class="delete-question-btn px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete Question</button>
        </div>
      `;
        questionList.insertAdjacentHTML('beforeend', questionHTML);
        updateDeleteButtons();
    });

    function updateDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', function () {
                button.parentElement.remove();
                toggleDeleteButtons();
            });
        });
        toggleDeleteButtons();
    }

    function toggleDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        if (deleteButtons.length === 0) {
            document.querySelectorAll('.delete-question-btn').forEach(button => {
                button.classList.add('hidden');
            });
        } else {
            document.querySelectorAll('.delete-question-btn').forEach(button => {
                button.classList.remove('hidden');
            });
        }
    }
    // Initial call to set up delete buttons
    updateDeleteButtons();
});
