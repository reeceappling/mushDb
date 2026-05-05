'use client'

import React, {useState} from "react";
import NotesAreaOld, {IsValidNote, Note, NotesAreaInline} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data, OnViewCreatorQuadCol, SplitAllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {AddToTransfers, InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import GenerationArea from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayInput,
    DisposedSaleContamArea,
    HandleJsonResponse,
    HandleTxtResponse,
    ImportDisplayInput,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems,
    NewEntryInput,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolveContamsFormData,
    resolvePicsFormData, SendMultipartRequest, setFormData,
    setFormImages, SingleListProps,
} from "@/app/components/common";
import ReaderWriterSelector from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensInlineDisplay, GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay,
    SpeciesArea, SubspeciesArea
} from "@/app/components/formSubcomponents/commonClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, InitialNotesState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {StasisTubeData} from "@/app/components/stasisTubeServer";
import {OnViewCreatorsQuadColArea, PcRunArea} from "@/app/components/pcRunClient";
import {PcRunData, RecentPCRunSelector} from "@/app/components/pcRunServer";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ExistingSpeciesSelector} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {SlantData} from "@/app/components/slantServer";
import {dataFor, InlineEntry} from "@/app/components/agarRecipeClient";
import {OvcForXfers} from "@/app/components/bagClient";
import {DisplayFormWrapper, ImportEntryFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {CreatedUpdatedDisposedArea} from "@/app/components/plateClient";
import {SpeciesSubspeciesArea} from "@/app/components/lcClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";

export function AssertStasisTube(input: any): asserts input is StasisTubeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('StasisTube assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['pcRun', 'string'],
        ['waterSource', 'string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('StasisTube assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function StasisTubeImportDisplay({headerLevel,cookies}:ImportDisplayInput) { // TODO: use headerLevel
    const [created, setCreated] = useState<number>(Date.now())
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>(undefined)
    const [generation, setGeneration] = useState<number | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    const importEntry = () => {
        let formData = new FormData()
        let dataObj: any = {
            created:created,
            //perms: perms, // TODO: validate on insert
        }
        if(species===undefined){
            setErr("Species must be set!")
            return
        }
        dataObj.species = species._id
        subspecies && (dataObj.subspecies = subspecies._id)
        knownFruitable && (dataObj.knownFruitable = knownFruitable)
        generation && (dataObj.generation = generation)
        if(imageFile!==undefined){
            formData.set("image", imageFile, "imgFile")
        }
        writeTagTo && (dataObj.writeTagTo=writeTagTo)

        SendMultipartRequest(BaseExternalUrl+"/db/import/stasisTube", cookies, formData)
            .then(HandleTxtResponse)
            .then((newId) => {
                redirect(BaseExternalUrl+"/view/stasisTube/"+newId)
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"stasisTube"}>
        {err!=undefined && <div>{"Error: "+err}</div>}
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies/*cookies={cookies}*/}/>
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies/*cookies={cookies}*/}/>
        <KnownFruitableArea doSelect={setKnownFruitable}/>
        <GenerationArea readonly={false} updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        {/*<EntryPermsArea setEntryPerms={setPerms}/>*/}
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />
        <button className={"greenButton"} onClick={importEntry}>{"Import Stasis Tube"}</button>
    </ImportEntryFormWrapper>
}

export default function StasisTubeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertStasisTube(data)
        const [initial, setInitial] = useState(data)
        const existingNotes: Note[] = initial.notes || []
        const initNotes: Data<Note>[] = existingNotes.map((n) => {
            return {data: n, disabled: false}
        })

        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [sale, setSale] = useState(initial.sale)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing:initNotes,new:[]})
        // State helpers
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: StasisTubeData)=>{
            setInitial(updated)
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setAcl(updated.acl)
        }
        const stasisTubeSubmit = () => {
            let body = new FormData()
            let dataObj:any={
                knownFruitable: knownFruitable,
                sale: sale,
                disposed: disposed,
                notes: notes,
                acl: acl,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(images)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                //body.set("data",JSON.stringify(dataObj)) // TODO: REDO THINGS ON GO SIDE (ENSURE OTHER PICTURE ONES DO THE SAME!
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl+"/db/update/stasisTube/"+initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertStasisTube(entry)
                    updateInitial(entry)
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: anything here?
        ]
        return (
            <DisplayFormWrapper entryType={"stasisTube"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Stasis Tube"} entryType={"stasisTube"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>{/* TODO: where to put?*/}
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel} />
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed} setDisposedOnParent={setDisposed} readonly={readonly}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <PcRunArea binaryId={initial.pcRun} headerLevel={headerLevel}/>
                        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly} headerLevel={headerLevel}/>
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel} canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore} headerLevel={headerLevel} />

                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        {/*<SpeciesSubspeciesFormArea species={initial.species} subspecies={initial.subspecies} />*/}
                        <InnocDisplay innoc={initial.innoc} openInNewTab={false}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"stasisTube"} allowNewTransferCreation={!readonly} transfersOut={transfersOut} validTypesTo={["plate","stasisTube","jar"/* TODO: ANYMORE????*/]} cookies={cookies} headerTxt={"Transfers"}/>
                <PicsDisplay pix={images} updateParent={setImages} readonly={readonly} headerLevel={headerLevel} />{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} current={contams} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    stasisTubeSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: StasisTube data format incorrect: " + err}</div>
    }
}

export function NewStasisTubeForm({handlers, pcRunIn}: {handlers: NewEntryInput<StasisTubeData>, pcRunIn?: PcRunData}){
    const [pcRun, setPcRun] = useState<PcRunData | undefined>(pcRunIn)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle isTopLevel
    const createStasisTube = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(pcRun===undefined){
            setErr("pc run must be defined")
            return
        }
        let body:any={pcRun: pcRun._id}
        writeTagTo && (body.writeTagTo = writeTagTo)

        fetch(BaseExternalUrl+"/create/stasisTube", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
            body: JSON.stringify(body),
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertStasisTube(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"stasisTube"}>
        <ErrorDisplay err={err} />
        {pcRunIn !== undefined && <RecentPCRunSelector doSelect={setPcRun} allowCreation={handlers.isTopLevel} creatorInPage={true}/>/* TODO: isTopLevel*/}
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} />{/*TODO: RFID SELECTOR*/}
        <button className={"greenButton"} onClick={createStasisTube}>{"Create"}</button>
    </NewEntryFormWrapper>

}

export function StasisTubeInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<StasisTubeData>) {
    const [expanded, setExpanded] = useState(expandByDefault)
    const b58id = data._id
    return <InlineEntry onClick={onClick}>
        <InlineSubArea props={{}}>
            <ID id={b58id} txt={"Stasis Tube"} entryType={"stasisTube"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
            <SpeciesArea readonly={true} initial={data.species}/>
            <SubspeciesArea readonly={true} currentSpecies={data.species} initialSub={data.subspecies} />
            <KnownFruitableArea initial={data.knownFruitable} readonly={true}/>
            <GensInlineDisplay gensSinceSpore={data.genSpore} gensSinceFruitOrSpore={data.genFruitOrSpore}/>
            <DisposedSaleContamArea sale={data.sale} disposed={data.disposed} contams={data.contamination}/>
        </InlineSubArea>
        <InlineExpansionArea props={{expanded:expanded}}>
            <PcRunArea binaryId={data.pcRun} offset={-1}/>
            {/*TODO: <ProjectsArea allowCreate={false} projects={data.perms?.projectPerms.ids} readonly={true} headerLevel={headerLevel} offset={-1} allowRemove={false}/>*/}
            <NotesAreaInline notes={data.notes} offset={-1}/>
            <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
        </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
                               expanded={expanded}/>
    </InlineEntry>
}

// export function StasisTubeListDisplay({data, onClick}: SingleListProps<StasisTubeData>) {
//     return <div>
//         {data.map((b,i)=>{
//             return <StasisTubeInline data={b} onClick={()=>{onClick(b)}} key={i}/>
//         })}
//     </div>
// }

export function StasisTubeListPageTable({data, onClick}: ListPageItems<StasisTubeData>) {
    const cols: ListTableColumn<StasisTubeData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}