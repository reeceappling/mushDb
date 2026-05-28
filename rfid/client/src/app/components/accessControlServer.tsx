export interface ACL {
    users?: Map<string,boolean>
    projects?: Map<string,boolean>
    blanketPerm?: boolean
}