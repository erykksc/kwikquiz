document.addEventListener('DOMContentLoaded', function() {
    const addQuestionBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');

    addQuestionBtn.addEventListener('click', function() {
        const questionNumber = document.querySelectorAll('.question-item').length + 1;
        const questionHTML = `
        <div class="question-item mb-4 p-4 border border-baby-pink rounded-lg">
          <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
          <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" required>
          <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
          <div class="answers-container answers-container-${questionNumber}">
            <!-- Answer Options will be added dynamically here -->
          </div>
          <!-- Add Answer Button -->
          <button type="button" class="add-answer-btn px-4 py-2 bg-dark-green text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green mb-2">Add Answer</button>
          <button type="button" class="delete-question-btn px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete Question</button>
        </div>
        `;
        questionList.insertAdjacentHTML('beforeend', questionHTML);
        updateDeleteButtons();
    });

    function updateDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', function() {
                button.parentElement.remove();
                updateQuestionNumbers();
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

    function updateQuestionNumbers() {
        const questionItems = document.querySelectorAll('.question-item');
        questionItems.forEach((item, index) => {
            const questionNumber = index + 1;
            const questionLabel = item.querySelector('label');
            questionLabel.textContent = `Question ${questionNumber}`;
            const inputs = item.querySelectorAll('input, textarea');
            inputs.forEach(input => {
                const name = input.name.replace(/-\d+/, `-${questionNumber}`);
                input.name = name;
                const id = input.id.replace(/-\d+/, `-${questionNumber}`);
                input.id = id;
                const label = item.querySelector(`label[for="${id}"]`);
                if (label) {
                    label.setAttribute('for', id);
                }
            });

            // Update answers-container class
            const answersContainer = item.querySelector(`.answers-container-${questionNumber}`);
            if (answersContainer) {
                answersContainer.classList = `answers-container answers-container-${questionNumber}`;
            }
        });
    }

    // Function to handle answer type change
    function handleAnswerTypeChange(event) {
        const answerInputName = event.target.getAttribute('data-answer');
        const answerType = event.target.getAttribute('data-type');
        const answerInput = document.querySelector(`input[name="${answerInputName}"], textarea[name="${answerInputName}"]`);

        if (!answerInput) return;

        switch (answerType) {
            case 'text':
                answerInput.outerHTML = `<input type="text" name="${answerInputName}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Enter text" required>`;
                break;
            case 'image':
                answerInput.outerHTML = `<input type="file" name="${answerInputName}" accept="image/*" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" required>`;
                break;
            case 'latex':
                answerInput.outerHTML = `<textarea name="${answerInputName}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Enter LaTeX" required></textarea>`;
                break;
        }
    }

    // Event delegation to handle answer type button clicks
    questionList.addEventListener('click', function(event) {
        if (event.target.classList.contains('add-answer-type-btn')) {
            handleAnswerTypeChange(event);
        }
    });

    // Event listener for adding answers dynamically
    questionList.addEventListener('click', function(event) {
        if (event.target.classList.contains('add-answer-btn')) {
            const questionItem = event.target.closest('.question-item');
            const questionNumber = questionItem.querySelector('label').textContent.trim().replace('Question ', '');
            const answersContainer = questionItem.querySelector(`.answers-container-${questionNumber}`);
            const answerCount = answersContainer.querySelectorAll('.answer-option').length + 1;
            const newAnswerHTML = `
              <div class="answer-option mb-2">
                <input type="text" name="answer-${questionNumber}-${answerCount}" class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2" placeholder="Option ${answerCount}" required>
                <div class="flex justify-between items-center mt-2">
                    <div>
                        <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="text">Text</button>
                        <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="image">Image</button>
                        <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerCount}" data-type="latex">LaTeX</button>
                    </div>
                    <div>
                        <button type="button" class="correct-answer-btn text-sm px-2 py-1 bg-red-500 text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-green-500">Incorrect</button>
                        <button type="button" class="delete-answer-btn text-sm px-2 py-1 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete</button>
                    </div>
                </div>
              </div>
            `;
            answersContainer.insertAdjacentHTML('beforeend', newAnswerHTML);
        }
    });

    // Event listener for deleting answers dynamically
    questionList.addEventListener('click', function(event) {
        if (event.target.classList.contains('delete-answer-btn')) {
            event.target.closest('.answer-option').remove();
        }
    });

    // Event listener for toggling correct/incorrect answer status
    questionList.addEventListener('click', function(event) {
        if (event.target.classList.contains('correct-answer-btn')) {
            const correctBtn = event.target;
            correctBtn.classList.toggle('bg-red-500');
            correctBtn.classList.toggle('bg-dark-green');
            correctBtn.classList.toggle('text-white');
            correctBtn.classList.toggle('text-white');
            correctBtn.textContent = correctBtn.textContent === 'Incorrect' ? 'Correct' : 'Incorrect';

            // Update hidden input
            const questionNumber = correctBtn.closest('.question-item').querySelector('label').textContent.trim().replace('Question ', '');
            const answerNumber = Array.from(correctBtn.closest('.answer-option').parentNode.children).indexOf(correctBtn.closest('.answer-option')) + 1;
            const hiddenInput = document.createElement('input');
            hiddenInput.type = 'hidden';
            hiddenInput.name = `correct-answer-${questionNumber}-${answerNumber}`;
            hiddenInput.value = correctBtn.textContent;
            correctBtn.parentNode.appendChild(hiddenInput);
        }
    });
});
