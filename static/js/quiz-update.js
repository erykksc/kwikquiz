document.addEventListener('DOMContentLoaded', function() {
    const addQuestionBtn = document.getElementById('add-question-btn');
    const questionList = document.getElementById('questions-list');

    // Parse quiz data from embedded JSON
    const quizData = JSON.parse(JSON.parse(document.getElementById('quiz-data').textContent));

    document.getElementById('title').value = quizData.Title;
    document.getElementById('password').value = quizData.Password;
    document.getElementById('description').value = quizData.Description;
    if (quizData.QuestionOrder === 'fixed') {
        document.getElementById('order-fixed').checked = true;
    } else {
        document.getElementById('order-random').checked = true;
    }

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
            <div class="question-item mb-4 p-4 border rounded-lg">
                <label class="block text-gray-700 font-semibold mb-2">Question ${questionNumber}</label>
                <input type="text" name="question-${questionNumber}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Enter question text" value="${question ? question.Text : ''}" required>
                <label class="block text-gray-700 font-semibold mb-2">Answer Options</label>
                ${[1, 2, 3, 4].map(i => `
                    <input type="text" name="answer-${questionNumber}-${i}" class="w-full px-4 py-2 mb-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" placeholder="Option ${i}" value="${question ? question.Answers[i - 1] : ''}" required>
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
