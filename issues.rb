require 'octokit'
require 'fileutils'

def get_open_issues_with_label(label)
  client = Octokit::Client.new(:access_token => ENV['GITHUB_PAT'])
  begin
    client.issues('thansetan/thansetan.github.io', :user => 'thansetan', :state => 'open', :per_page => 100, :labels => label)
  rescue Octokit::NotFound
    nil
  end
end

def generate_filename(issue)
  date = issue.updated_at.strftime("%Y-%m-%d")
  title = issue.title.downcase.gsub(' ', '-').gsub(/[^\w-]/, '')
  "#{date}-#{title}.md"
end

def update_posts_directory(issues, dir)
  FileUtils.mkdir_p(dir) unless Dir.exist?(dir)
  
  issue_filename_list = issues.map { |issue| generate_filename(issue) }
  local_files = Dir.glob("#{dir}/*.md").map { |file| File.basename(file) }

  local_files.each do |file_name| # if local file is not in the list of issues, delete it
    File.delete(dir + '/' + file_name) unless issue_filename_list.include?(file_name)
  end
  
  issues.each do |issue| # list all issues and create a file for each issue
    file_name = generate_filename(issue)
    File.open(dir + '/' + file_name, 'w') { |file| file.write(issue.body) }
  end
end

dir = 'contents/_posts'
issues = get_open_issues_with_label('published')
update_posts_directory(issues, dir) if issues
