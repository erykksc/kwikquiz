document.addEventListener('DOMContentLoaded', function() {
    const addQuestionBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');

    const titleElement = document.getElementById('quiz-title');
    const headingElement = document.getElementById('quiz-heading');
    const submitButton = document.getElementById('submit-btn');
    const deleteButton = document.getElementById('delete-btn');
    const quizForm = document.getElementById('quiz-form_1');

    // Parse quiz data from embedded JSON
    const quizDataElement = document.getElementById('quiz-data');
    const quizData = quizDataElement ? JSON.parse(JSON.parse(quizDataElement.textContent)) : null;

    if (quizData) {
        // Dynamically change labels and function of buttons
        titleElement.textContent = 'Edit Quiz';
        headingElement.textContent = 'Edit KWIKQUIZ';
        submitButton.textContent = 'UPDATE KWIKQUIZ';
        quizForm.dataset.hxPut = `/quizzes/update/${quizData.ID}`;
        deleteButton.classList.remove('hidden');

        // automatically fill in the forms
        document.getElementById('title').value = quizData.Title;
        document.getElementById('password').value = quizData.Password;
        document.getElementById('description').value = quizData.Description;

        // Add existing questions
        quizData.Questions.forEach((question, index) => {
            addQuestion(question, index + 1)
        });
    }else {
        titleElement.textContent = 'Create Quiz';
        headingElement.textContent = 'Create a new KWIKQUIZ';
        submitButton.textContent = 'CREATE KWIKQUIZ';
        quizForm.setAttribute('hx-post', '/quizzes/create/');
    }

    addQuestionBtn.addEventListener('click', function() {
        const questionNumber = document.querySelectorAll('.question-item').length + 1;
        addQuestion(null, questionNumber);
    });

    function addQuestion(question, questionNumber) {
        const questionHTML = `
            <div class="question-item mb-4 p-4 border rounded-lg">
                <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
                <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" value="${question ? question.Text : ''}" required>
                <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
                ${[1, 2, 3, 4].map(i => `
                    <input type="text" name="answer-${questionNumber}-${i}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option ${i}" value="${question ? question.Answers[i - 1].Text : ''}" required>
                `).join('')}
                <div class="mb-4">
                    <label for="correct-answer-${questionNumber}" class="block text-gray-700 font-semibold mb-2">Correct Answer</label>
                    <div class="flex items-center">
                        ${[1, 2, 3, 4].map(i => `
                            <input type="radio" id="option-${questionNumber}-${i}" name="correct-answer-${questionNumber}" value="${i}" class="mr-2" ${question && question.CorrectAnswer === i ? 'checked' : ''} required>
                            <label for="option-${questionNumber}-${i}" class="mr-4 text-gray-700">Option ${i}</label>
                        `).join('')}
                    </div>
                </div>
                <button type="button" class="delete-question-btn px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 focus:outline-none focus:ring-2 focus:ring-red-500">Delete Question</button>
            </div>
        `;
        questionList.insertAdjacentHTML('beforeend', questionHTML);
        updateDeleteButtons();
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

