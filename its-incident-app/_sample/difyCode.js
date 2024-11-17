function main({ answer1, answer2, answer3 }) {
  return {
    result: [{ answer1: answer1 }, { answer2: answer2 }, { answer3: answer3 }],
  };
}

function main(answer) {
  return answer.map((str, index) => {
    return {
      [`answer${index + 1}`]: str,
    };
  });
}
