'use client'

import React, {ChangeEvent, JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {
    clientPostRequestHeaders,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey
} from "@/app/components/common";
import {ErrorDisplay, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import {ProjectData,} from "@/app/components/projectServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ProjectWithPerm} from "@/app/components/perms";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {IsStringMapToString} from "@/app/components/accessControlClient";
import {HandleErr, UserSelector} from "@/app/components/userClient";
import TestAndValidate from "@/app/components/testing/untested";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertProject(input: any): asserts input is ProjectData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Project assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['completed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Project assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['perms', IsStringMapToString], // TODO: THIS IS NOT WORKING PROPERLY!!! CHANGE TO STRING FORMAT!!!!
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Project assertion failure: required key ' + key + ' was not valid. was ' + JSON.stringify(input[key]));
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function ProjectDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput) {
    try {
        AssertProject(data)
        const [initial, setInitial] = useState(data)
        const initPerms = new Map<string, string>(Object.entries(data.perms || {}) as [string, string][]) // TODO: IF THIS WORKS USE IT FOR UNMARSHALLING ALL PARMS!

        const [completed, setCompleted] = useState(data.completed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [perms, setPerms] = useState<Map<string, string>>(initPerms)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: ProjectData) => {
            setInitial(updated)
            const ps = new Map<string, string>(Object.entries(updated.perms || {}) as [string, string][]) // TODO: IF THIS WORKS USE IT FOR UNMARSHALLING ALL PARMS!

            setCompleted(updated.completed)
            setNotes(InitialNotesState(updated.notes))
            setPerms(ps)
            setErr(undefined)
        }
        // TODO: users and entries for this!
        const completedArea = () => {
            const handleCompletedClick = () => {
                setCompleted(completed ? undefined : Date.now())
            }
            if (readonly || initial.completed) {
                let isComp = "In-Progress"
                if (completed) {
                    isComp = "Completed " + NumberToDate(new Date(completed))
                }
                return <div>
                    <div>{isComp}</div>
                </div>
            }
            return <div>
                <div>{"Completed: "}</div>
                <input type={"checkbox"} checked={!!completed} onChange={() => {
                    // TODO: ensure onChange does not need anything
                }} onClick={handleCompletedClick} onSubmit={(e) => {
                    e.preventDefault();
                }}/>
            </div>
        }
        const cookies = useContext(CookiesContext)
        const projectSubmit = () => {
            const body: any = {
                notes: notes,
                completed: completed,
                perms: Object.fromEntries(perms), // TODO: ensure this is being done on any maps that are being marshalled!!!!!
            }
            console.log("sending perms: " + JSON.stringify(Object.fromEntries(perms)))


            DoUpdateRequest("project",encodeURIComponent(data._id), body, AssertProject, allCookies(cookies))
                .then(v=>{
                    updateInitial(new ProjectData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        return (
            <DisplayFormWrapper entryType={"project"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                {/* data._id on next line because project name can have spaces?*/}
                <ID id={data._id} txt={"Project"} entryType={"project"}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Created: "} when={initial.creationDate} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        {completedArea()}
                    </FlexedSinglesGroup>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TestAndValidate todos={["setting a user to view only and updating will remove the user from the project :("]}>{/* TODO: THIS*/}
                    <ProjectPermsArea perms={perms} setPerms={setPerms}
                                      readonly={readonly}/> {/* TODO: HEAVILY TEST! Also ensure this is properly covered on the go side!*/}
                </TestAndValidate>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    projectSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Plate data format incorrect: " + err}</div>
    }
}

export function NewProjectForm(
    {handlers}: { handlers: NewEntryInput<ProjectData> }) { // TODO: add cookies?
    const [name, setName] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    // TODO: load up user on server side into the userperms as write (unless blanket is write)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle isTopLevel

    const cookies = useContext(CookiesContext)
    const createProject = () => {
        if (name === undefined) {
            setErr("Name field cannot be undefined")
            return
        }
        const body = {
            name: name, // TODO: validate that project name is valid for url
            notes: notes,
        }
        DoCreateRequest("project", body, AssertProject, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    const handleChangeName = (event: ChangeEvent<HTMLInputElement>) => {
        setName(event.target.value)
    }
    const updateNotes = (entries: AllEntries<Note>) => {
        setNotes(entries.new.map((e) => {
            return e.data
        }))
    }
    const projectNameArea = () => {
        return <div>
            <InputTextInlineTitle label={"Project Name"} readonly={false} value={name || ""} onChange={setName} />
        </div>
    }
    return <NewEntryFormWrapper entryType={"project"}>
        <ErrorDisplay err={err}/>
        {projectNameArea()}
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={createProject}>{"Create Project"}</button>
    </NewEntryFormWrapper>

}

export function NumberToPerm(n?: number) {
    if (n === undefined || n === 0) {
        return undefined
    }
    return n === 2;
}

export function PermToNumber(p: boolean | undefined): number {
    if (p === undefined) {
        return 0
    }
    if (!p) {
        return 1
    }
    return 2
}

export function ReadWriteAdminSelector({readonly, onUpdate, value}: {
    value: string,
    readonly: boolean,
    onUpdate?: (valOut: string) => void
}) {
    const strForVal = (str?: string) => {
        return (value === "read") ? "can view" : (value === "admin" ? "is admin" : "can edit")
    }
    if (readonly) {
        return <text>{strForVal(value)}</text>
    }
    return <SelectorFor options={["can view", "can edit", "is admin"]} initial={strForVal(value)}
                        updateParent={s => {
                            if (s === "is admin") {
                                onUpdate && onUpdate("admin")
                            }
                            if (s === "can edit") {
                                onUpdate && onUpdate("write")
                            }
                            if (s === "can view") {
                                onUpdate && onUpdate("read")
                            }
                            return
                        }} disabled={false}/>
}

export function ReadWriteSelector({readonly, onUpdate, value}: {
    value: boolean,
    readonly: boolean,
    onUpdate?: (b: boolean) => void
}) {
    if (readonly) {
        return <text>{value ? "can edit" : "can view"}</text>
    }
    return <SelectorFor options={["can view", "can edit"]} initial={value ? "can edit" : "can view"}
                        updateParent={s => {
                            onUpdate && onUpdate(s === "can edit")
                        }} disabled={false}/>
}

export function ProjectPermsArea({perms, setPerms, readonly}: {
    perms?: Map<string, string>, // TODO: to string!
    setPerms?: (pp: Map<string, string>) => void, // TODO: to string!
    readonly: boolean,
}) {
    const depth = useContext(DepthContext)
    if (readonly && !perms) {
        return null
    }
    return <DepthProvider>
        <div className={"subForm depth" + depth}>
            <div className={"centerH text-lg mb-1"}>{"Permissions"}</div>
            <div className={"projectPermsUsers"}>
                {/* TODO: make this into a grid or table*/}
                {perms !== undefined && perms.size > 0 && [...perms.entries()].map((p) => { // TODO: how to handle the undefineds?
                    return <>
                        <div key={p[0] + "name"}>{p[0]}</div>
                        <ReadWriteAdminSelector key={p[0] + "sel"} readonly={readonly} value={p[1]}
                                                onUpdate={(b) => {
                                                    setPerms && setPerms(new Map(perms).set(p[0], b))
                                                }}/>
                        <RemoveButton key={p[0] + "remv"} click={() => {
                            const updated = new Map<string, string>()
                            perms.entries().forEach(v => {
                                if (v[0] !== p[0]) {
                                    updated.set(v[0], v[1])
                                }
                            })
                            setPerms && setPerms(updated)
                        }} txt={"Remove"}/>
                    </>
                })}
            </div>

            {/* AREA TO ADD USER */}
            <div className={"inlineChildren"}>{"Add user: "}<UserSelector onSelect={(u) => {
                const out = structuredClone(perms) || new Map<string, string>()
                out.set(u._id, "read")
                setPerms && setPerms(out)
            }} blacklist={(perms !== undefined && perms.size > 0) ? [...perms.entries()].map(u => {
                return u[0]
            }) : []}/>
            </div>
        </div>
    </DepthProvider>
}

// // TODO: FIX FOR NEW PROJECT PERMS!
// export function ProjectPermsAreaOLD({originalPerms, setPerms, canEdit}: { // TODO: this whole thing?
//     originalPerms?: ProjectPerms,
//     setPerms?: (pp: ProjectPerms) => void,
//     canEdit: boolean,
// }) {
//     const [fullPerms, setFullPerms] = useState<ProjectPerms>(originalPerms || {
//         users: {ids: [], canWrite: []},
//         blanket: 0
//     })
//     const [userToAdd, setUserToAdd] = useState<string>("")
//     const [userToAddCanWrite, setUserToAddCanWrite] = useState<boolean>(false)
//     const [err, setErr] = useState<string | undefined>()
//     //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     if (!canEdit) {
//         const blanketText = () => {
//             switch (fullPerms.blanket) {
//                 case 0:
//                     return ""
//                 case 1:
//                     return "Read"
//                 case 2:
//                     return "Read, Write"
//             }
//             if (fullPerms.blanket === 0) {
//                 return "No blanket permissions"
//             }
//         }
//         const textForCanWrite = (canWrite: boolean) => {
//             if (canWrite) {
//                 return "Write"
//             }
//             return "Read"
//         }
//         return <div>
//             <div>{"Blanket permissions: " + blanketText()}</div>
//             <div>
//                 <div>{"User permissions"}</div>
//                 <table>
//                     <tr>
//                         <th>{"User"}</th>
//                         <th>{"Permission"}</th>
//                     </tr>
//                     {fullPerms.users.ids.map((user, i) => {
//                         return <tr key={i}>
//                             <td>{user.val}</td>
//                             <td>{textForCanWrite(fullPerms.users.canWrite[i])}</td>
//                         </tr>
//                     })}
//                 </table>
//             </div>
//         </div>
//     }
//
//
//     const doUpdate = (newPerms?: ProjectPerms) => {
//         let tempPerms = newPerms || {users: {ids: [], canWrite: []}, blanket: 0}
//         if (newPerms) {
//             tempPerms = newPerms
//         }
//         setFullPerms(tempPerms)
//         if (!setPerms) {
//             return
//         }
//         // remove unnecessary perms as needed
//         if (tempPerms.blanket === 0) {
//             setPerms(tempPerms)
//             return
//         }
//         let toSet: ProjectPerms = {users: {ids: [], canWrite: []}, blanket: tempPerms.blanket}
//         if (tempPerms.blanket === 1) {
//             for (let i = 0; i < tempPerms.users.ids.length; i++) {
//                 if (tempPerms.users.canWrite[i]) {
//                     toSet.users.ids.push(tempPerms.users.ids[i])
//                     toSet.users.canWrite.push(true)
//                 }
//             }
//         }
//         setPerms(toSet)
//     }
//
//     const changeBlanket = (b?: boolean) => {
//         let updated = {...fullPerms}
//         updated.blanket = PermToNumber(b)
//         doUpdate(updated)
//     }
//     const changeUser = (un: number) => {
//         return (newBp?: boolean) => {
//             let updated = {...fullPerms}
//             if (newBp === undefined) {
//                 // remove that user
//                 const filterFunc = (_: any, index: number) => {
//                     return index !== un
//                 }
//                 updated.users.ids = [...fullPerms.users.ids].filter(filterFunc);
//                 updated.users.canWrite = [...fullPerms.users.canWrite].filter(filterFunc);
//
//             } else {
//                 // Update that user
//                 updated.users.canWrite[un] = newBp
//             }
//             doUpdate(updated)
//         }
//     }
//     const addUserByEmailOrUsername = (val: string) => {
//         if (setPerms) {
//             fetch(BaseExternalUrl + "/db/userIdFor", { // TODO: handle on the go side
//                 method: 'Get',
//                 body: val,
//                 headers: clientPostRequestHeaders,
//             }).then(HandleTxtResponse).then((id) => {
//                 let updated = {...fullPerms}
//                 updated.users.ids.push({id: id, val: val})
//                 updated.users.canWrite.push(userToAddCanWrite)
//                 doUpdate(updated)
//             }).catch((e: string) => {
//                 setErr(e)
//             })
//         }
//     }
//     return <div>
//         <div>
//             {"Global Permission: "}<PermissionSelector onChange={changeBlanket}
//                                                        canWrite={NumberToPerm(fullPerms.blanket)}/>
//         </div>
//         {/* TODO: REMOVE ALL UNNECESSARY USERS IF BLANKET CHANGE SHOULD*/}
//         <div>
//             <div>{/* Current Users */}
//                 {fullPerms.users.ids.map((uid: UserIdPair, i: number) => {
//                     return <div className={"projPermsAreaUser"} key={i}>
//                         {uid.val + ": "}<PermissionSelector onChange={changeUser(i)}
//                                                             canWrite={fullPerms.users.canWrite[i]}/>
//                     </div>
//                 })}
//             </div>
//             <div>
//                 <ErrorDisplay err={err}/>
//                 {/* Add new user by email */}
//                 <Textbox readonly={false} label="New user email or username" value={userToAdd} fieldName={"FIXME"}
//                          updateTextHandler={setUserToAdd}/>
//                 <PermissionSelector onChange={b => {
//                     b && setUserToAddCanWrite(b)
//                 }} canWrite={userToAddCanWrite}
//                                     dontShowBelow={fullPerms.blanket === 0 ? false : NumberToPerm(fullPerms.blanket)}/>
//                 <button onClick={() => {
//                     addUserByEmailOrUsername(userToAdd)
//                 }}> {"ADD USER"}</button>
//             </div>
//         </div>
//         <button onClick={() => {
//             doUpdate(originalPerms || {users: {ids: [], canWrite: []}, blanket: 0})
//         }}>{"Revert permission changes"}</button>
//     </div>
// }

export async function GetMyProjects() {
    return await fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as ProjectWithPerm[]
        } catch (err) {
            throw err
        }
    })
}

// TODO: used to return Promise<ProjectWithPerm[]>
export async function GetSessionUserProjects(complete?:boolean): Promise<string[]> {
    let params = ""
    if (complete !== undefined) {
        params = "?complete="+(complete?"true":"false")
    }
    return fetch(BaseExternalUrl + "/sessionUserProjects"+params, { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as string[]
        } catch (err) {
            throw err
        }
    })
}

export function ProjectsSelector(inp: {
    onSelect: (projName: string) => void
    blacklist?: string[]
    complete?: boolean // TODO: use?
}) {
    const [loading, setLoading] = useState(true)
    const [projects, setProjects] = useState<string[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        GetSessionUserProjects(inp.complete).then(projNames=>{
            setProjects(projNames)
            setErr(undefined)
            setLoading(false)
        }).catch(e=>{
            HandleErr(e, setErr)
        })
        // fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: ensure user has projects attached
        //     // TODO: do we also want to pull the user's perms for each project?
        //     method: "GET",
        //     headers: clientPostRequestHeaders,
        // }).then(HandleJsonResponse).then((projs) => {
        //     try {
        //         return projs as string[] // TODO: FIXME!
        //     } catch (err) {
        //         throw err
        //     }
        // }).then(projs => {
        //     setProjects(projs)
        //     setLoading(false)
        //     setErr(undefined)
        //     return
        // }).catch(err => {
        //     HandleErr(err, setErr)
        //     return
        // })
    }, [])
    if (loading) {
        return <div>{"Loading projects selector"}</div>
    }
    const projectOptions = () => {
        if (projects == undefined || projects.length == 0) {
            return [""]
        }
        return ["", ...projects.filter(pToFilter => {
            return (inp.blacklist || []).indexOf(pToFilter) == -1
        })]
    }
    return <div>
        <ErrorDisplay err={err}/>
        <SelectorResetsOnSelectFor options={projectOptions()} updateParent={(pr) => {
            inp.onSelect(pr)
        }}/>
    </div>
}

export function ProjectListPageTable({data, onClick, withLink}: ListPageItems<ProjectData>) {
    let cols: ListTableColumn<ProjectData>[] = [
        NewColumn("Name", (v) => v._id),
        NewColumn("Completed", (v) => {
            return v.completed ? NumberToDateStr(v.completed) : ""
        }),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: ProjectData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new ProjectData(v)}}/>
}

export function ProjectSelectorTable({data, onClick}: ListPageItems<ProjectData>) {
    return <ProjectListPageTable data={data} onClick={onClick} withLink={true} />
}

// TODO: distinguish from ProjectsSelector
export function ProjectSelector(
    {
        doSelect,
        allowCreate,
    }: {
        doSelect: (val: ProjectData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: ProjectData[]):JSX.Element=>{
        return <ProjectSelectorTable data={items} onClick={doSelect}/>
    }
    const creationArea = ()=>{
        if(!allowCreate){
            return null
        }
        return <NewProjectForm handlers={{onCreate: doSelect,isTopLevel: false}}/>
    }

    return <ExistingRecentSelector entryType={"project"} entryTypes={"projects"} doSelect={doSelect} asserter={AssertProject} table={table}>
        {creationArea()}
    </ExistingRecentSelector>
}