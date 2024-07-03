document.addEventListener('DOMContentLoaded', function() {
    const addQuestionBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');

    // Parse quiz data from embedded JSON
    const quizData = JSON.parse(JSON.parse(document.getElementById('quiz-data').textContent));

    document.getElementById('title').value = quizData.Title;
    document.getElementById('password').value = quizData.Password;
    document.getElementById('description').value = quizData.Description;

    // Add existing questions
    quizData.Questions.forEach((question, index) => {
        addQuestion(question, index + 1);
    });

    addQuestionBtn.addEventListener('click', function() {
        const questionNumber = document.querySelectorAll('.question-item').length + 1;
        addQuestion(null, questionNumber);
    });

    function addQuestion(question, questionNumber) {
        const questionHTML = `
        <div class="question-item mb-4 p-4 border border-baby-pink rounded-lg">
          <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
          <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" value="${question ? question.Text : ''}" required>
          <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
          <div class="answers-container answers-container-${questionNumber}">
            ${question ? question.Answers.map((answer, index) => getAnswerHTML(questionNumber, index + 1, answer)).join('') : ''}
          </div>
          <!-- Add Answer Button -->
          <button type="button" class="add-answer-btn px-4 py-2 bg-dark-green text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green mb-2">Add Answer</button>
          <button type="button" class="delete-question-btn px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete Question</button>
        </div>
        `;
        questionList.insertAdjacentHTML('beforeend', questionHTML);
        updateDeleteButtons();
    }

    function getAnswerHTML(questionNumber, answerNumber, answer) {
        const isCorrect = answer && answer.IsCorrect;
        return `
          <div class="answer-option mb-2">
            <input type="${answer && answer.Image ? 'file' : 'text'}" name="answer-${questionNumber}-${answerNumber}"
                class="w-full px-4 py-2 border border-dark-green rounded-lg focus:outline-none focus:ring-2 focus:ring-dark-green mb-2"
                placeholder="Option ${answerNumber}"
                value="${answer ? (answer.Text || answer.LaTeX || '') : ''}"
                ${answer && answer.Image ? 'accept="image/*"' : ''} required>
            <div class="flex justify-between items-center mt-2">
                <div>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerNumber}" data-type="text">Text</button>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerNumber}" data-type="image">Image</button>
                    <button type="button" class="add-answer-type-btn text-sm px-2 py-1 bg-baby-pink-button text-dark-green rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-dark-green" data-answer="answer-${questionNumber}-${answerNumber}" data-type="latex">LaTeX</button>
                </div>
                <div>
                    <button type="button" class="correct-answer-btn text-sm px-2 py-1 ${isCorrect ? 'bg-dark-green' : 'bg-red-500'} text-white rounded-lg hover-bg-baby-pink focus:outline-none focus:ring-2 focus:ring-green-500">${isCorrect ? 'Correct' : 'Incorrect'}</button>
                    <button type="button" class="delete-answer-btn text-sm px-2 py-1 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete</button>
                </div>
            </div>
          </div>
        `;
    }

    function updateDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', function() {
                button.parentElement.remove();
                toggleDeleteButtons();
            });
        });
        toggleDeleteButtons();
    }

    function toggleDeleteButtons() {
        const deleteButtons = document.querySelectorAll('.delete-question-btn');
        if (deleteButtons.length === 1) {
            deleteButtons[0].classList.add('hidden');
        } else {
            deleteButtons.forEach(button => button.classList.remove('hidden'));
        }
    }

    function updateQuestionNumbers() {
        document.querySelectorAll('.question-item').forEach((item, index) => {
            const questionNumber = index + 1;
            item.querySelector('label').textContent = `Question ${questionNumber}`;
            item.querySelectorAll('input, textarea').forEach(input => {
                input.name = input.name.replace(/-\d+/, `-${questionNumber}`);
            });
            item.querySelector('.answers-container').className = `answers-container answers-container-${questionNumber}`;
        });
    }

    // Event delegation for dynamically added elements
    questionList.addEventListener('click', function(event) {
        if (event.target.classList.contains('add-answer-btn')) {
            const questionItem = event.target.closest('.question-item');
            const questionNumber = questionItem.querySelector('label').textContent.trim().replace('Question ', '');
            const answersContainer = questionItem.querySelector(`.answers-container-${questionNumber}`);
            const answerCount = answersContainer.querySelectorAll('.answer-option').length + 1;
            answersContainer.insertAdjacentHTML('beforeend', getAnswerHTML(questionNumber, answerCount));
        } else if (event.target.classList.contains('add-answer-type-btn')) {
            handleAnswerTypeChange(event);
        } else if (event.target.classList.contains('delete-answer-btn')) {
            event.target.closest('.answer-option').remove();
        } else if (event.target.classList.contains('correct-answer-btn')) {
            toggleCorrectAnswer(event.target);
        }
    });

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

    function toggleCorrectAnswer(button) {
        button.classList.toggle('bg-red-500');
        button.classList.toggle('bg-dark-green');
        button.textContent = button.textContent === 'Incorrect' ? 'Correct' : 'Incorrect';

        // Update hidden input
        const questionNumber = button.closest('.question-item').querySelector('label').textContent.trim().replace('Question ', '');
        const answerNumber = Array.from(button.closest('.answer-option').parentNode.children).indexOf(button.closest('.answer-option')) + 1;
        const hiddenInput = button.parentNode.querySelector(`input[name="correct-answer-${questionNumber}-${answerNumber}"]`) || document.createElement('input');
        hiddenInput.type = 'hidden';
        hiddenInput.name = `correct-answer-${questionNumber}-${answerNumber}`;
        hiddenInput.value = button.textContent === 'Correct';
        button.parentNode.appendChild(hiddenInput);
    }

    // Initial call to set up delete buttons
    updateDeleteButtons();
});
