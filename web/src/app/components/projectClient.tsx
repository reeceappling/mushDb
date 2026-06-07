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
import {IsStringMapToString, UnmarshalAclMapField} from "@/app/components/accessControlClient";
import {HandleErr, UserSelector} from "@/app/components/userClient";
import TestAndValidate from "@/app/components/testing/untested";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {ACL} from "@/app/components/accessControlServer";
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
    input.perms = UnmarshalProjectPermsField(input) // TODO: is this optional????
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        //['perms', IsStringMapToString], // TODO: THIS IS NOT WORKING PROPERLY!!! CHANGE TO STRING FORMAT!!!! (try above)
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

export function UnmarshalProjectPermsField(input: any): Map<string,string> {
   // TODO: needs to be able to throw!
    if ('perms' in input && (input.perms !== undefined)) { // TODO: ensure works properly, we don't want perms getting messed up
        const field = input.perms
        if (typeof field !== 'object' || field === null) {
            throw 'perms field must be an object'
        }
        // Exists, return populated map
        return new Map<string, string>(Object.entries(field as Record<string, string>))
    } else {
        // Does not exist or is undefined, return an empty map
        return new Map<string, string>()
    }
}

export default function ProjectDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput<ProjectData>) {
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
                perms: Object.fromEntries(perms), // TODO: ensure this works!
            }
            console.log("sending perms: " + JSON.stringify(Object.fromEntries(perms)))


            // TODO: Separate project perms request?
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
                    <ProjectPermsArea perms={perms} setPerms={setPerms} /* TODO: ENSURE ADDING/CHANGING USERS MAKES A SEPARATE REQUEST TO THE SERVER!*/
                                      readonly={readonly}/> {/* TODO: HEAVILY TEST! Also ensure this is properly covered on the go side!*/}
                </TestAndValidate>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    projectSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewProjectForm(
    {handlers}: { handlers: NewEntryInput<ProjectData> }) {
    const [name, setName] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
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
    return <NewEntryFormWrapper entryType={"project"}>
        <ErrorDisplay err={err}/>
        <div>
            <InputTextInlineTitle label={"Project Name"} readonly={false} value={name || ""} onChange={setName} />
        </div>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={createProject}>{"Create Project"}</button>
    </NewEntryFormWrapper>
}

// export function NumberToPerm(n?: number) {
//     if (n === undefined || n === 0) {
//         return undefined
//     }
//     return n === 2;
// }
//
// export function PermToNumber(p: boolean | undefined): number {
//     if (p === undefined) {
//         return 0
//     }
//     if (!p) {
//         return 1
//     }
//     return 2
// }

export function ReadWriteAdminSelector({readonly, onUpdate, value}: {
    value: string,
    readonly: boolean,
    onUpdate?: (valOut: string) => void
}) {
    const strForVal = (str?: string) => {
        return (str === "read") ? "can view" : (str === "admin" ? "is admin" : "can edit")
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
                            // TODO: what about the blank option???
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
    perms?: Map<string, string>,
    setPerms?: (pp: Map<string, string>) => void,
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

export async function GetMyProjects() { // TODO: use
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
            return projs as string[] // TODO: ensure this return is correct?
        } catch (err) {
            throw err
        }
    })
}

// ProjectsSelector shows shows only projects that a user has associated with them
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
// ProjectSelector shows ALL projects, not just projects that a user has associated with them // TODO: is this ok?
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